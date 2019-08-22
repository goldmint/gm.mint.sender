package pool

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/void616/gm-sumusrpc/conn"
)

// NodePool is a connections pool of any single node
type NodePool struct {
	endpoint           string
	connOpts           conn.Options
	closeFlag          *int32
	availableFlag      *int32
	maxConns           int32
	consumersCount     *int32
	consumedConnsCount *int32
	requestChan        chan struct{}
	consumerChan       chan *Conn
	releasedChan       chan *Conn
	routineWG          sync.WaitGroup
}

// newNodePool creates new NodePool instance
func newNodePool(endpoint string, copts conn.Options, maxConns int32) *NodePool {
	ret := &NodePool{
		endpoint:           endpoint,
		connOpts:           copts,
		closeFlag:          new(int32),
		availableFlag:      new(int32),
		maxConns:           maxConns,
		consumersCount:     new(int32),
		consumedConnsCount: new(int32),
		requestChan:        make(chan struct{}, 1),
		consumerChan:       make(chan *Conn),
		releasedChan:       make(chan *Conn),
		routineWG:          sync.WaitGroup{},
	}
	if ret.maxConns < 1 {
		ret.maxConns = 1
	}
	*ret.availableFlag = 1

	// launch routine
	ret.routineWG.Add(1)
	go ret.routine()

	return ret
}

// NotifyClose pool
func (p *NodePool) NotifyClose() {
	atomic.StoreInt32(p.closeFlag, 1)
}

// Close pool
func (p *NodePool) Close() error {
	atomic.StoreInt32(p.closeFlag, 1)
	p.routineWG.Wait()
	close(p.requestChan)
	close(p.consumerChan)
	close(p.releasedChan)
	return nil
}

// Available check
func (p *NodePool) Available() bool {
	return atomic.LoadInt32(p.availableFlag) != 0
}

// Get connection
func (p *NodePool) Get(timeout time.Duration) (*Conn, error) {

	if p.closing() {
		return nil, fmt.Errorf("Pool is closed")
	}

	atomic.AddInt32(p.consumersCount, 1)
	defer atomic.AddInt32(p.consumersCount, -1)

	// notify routine
	select {
	case p.requestChan <- struct{}{}:
	default:
	}

	select {
	case conn := <-p.consumerChan:
		if conn == nil {
			return nil, fmt.Errorf("Failed to get a connection")
		}
		return conn, nil

	case <-time.After(timeout):
		return nil, fmt.Errorf("Failed to get a connection (timeout)")
	}
}

// PendingConsumers count
func (p *NodePool) PendingConsumers() int32 {
	return atomic.LoadInt32(p.consumersCount)
}

// ConsumedConnections count
func (p *NodePool) ConsumedConnections() int32 {
	return atomic.LoadInt32(p.consumedConnsCount)
}

// closing checks node required to close
func (p *NodePool) closing() bool {
	return atomic.LoadInt32(p.closeFlag) != 0
}

// routine manages node's pool
func (p *NodePool) routine() {

	defer p.routineWG.Done()

	freeConns := list.New()
	leave := false

	onConnReleased := func(conn *Conn) {
		if !conn.holder.Dead() {
			freeConns.PushBack(conn)
		} else {
			go conn.holder.Close()
		}
		atomic.AddInt32(p.consumedConnsCount, -1)
	}

	for !(p.closing() && atomic.LoadInt32(p.consumersCount) == 0 && atomic.LoadInt32(p.consumedConnsCount) == 0) {

		// wait for event
		select {
		case _ = <-p.requestChan:
		case conn := <-p.releasedChan:
			onConnReleased(conn)
		case <-time.After(time.Millisecond * 5):
		}

		// try to move released connections to the pool
		leave = false
		for !leave {
			select {
			case conn := <-p.releasedChan:
				onConnReleased(conn)
			default:
				leave = true
			}
		}

		// try to feed consumers with connections from the pool
		leave = false
		for !leave {
			// pool still working
			if !p.closing() {
				e := freeConns.Front()
				if e == nil {
					leave = true
					continue
				}
				select {
				case p.consumerChan <- e.Value.(*Conn):
					freeConns.Remove(freeConns.Front())
					atomic.AddInt32(p.consumedConnsCount, 1)
				default:
					leave = true
				}
			} else {
				// pool is closed: feed with nil
				select {
				case p.consumerChan <- nil:
				default:
					leave = true
				}
			}
		}

		// a new connection can be allocated
		if !p.closing() {
			canAllocate := p.maxConns - atomic.LoadInt32(p.consumedConnsCount) - int32(freeConns.Len())
			if freeConns.Len() == 0 && canAllocate > 0 {
				iface, err := newConnHolder(p.endpoint, p.connOpts)
				if iface != nil && err == nil {
					conn := newConn(
						iface,
						p.releasedChan,
					)
					freeConns.PushBack(conn)

					atomic.StoreInt32(p.availableFlag, 1)
				} else {
					atomic.StoreInt32(p.availableFlag, 0)
				}
			}
		}

		// check dead connections
		if !p.closing() {
			for e := freeConns.Front(); e != nil; {
				x := e
				e = e.Next()

				if x.Value.(*Conn).holder.Dead() {
					freeConns.Remove(x).(*Conn).holder.Close()
				}
			}
		}
	}

	// release actual connections
	for e := freeConns.Front(); e != nil; e = e.Next() {
		x := e
		freeConns.Remove(x).(*Conn).holder.Close()
	}
}
