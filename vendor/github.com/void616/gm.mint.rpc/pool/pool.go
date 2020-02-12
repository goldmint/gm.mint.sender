package pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/void616/gm.mint.rpc/conn"
)

// Pool is a pool of pools (nodes) of connections
type Pool struct {
	closeFlag bool
	nodesLock sync.RWMutex
	nodes     map[string]*NodePool
	balancer  Balancer
}

// NodeMeta contains node's metadata for balancer
type NodeMeta struct {
	Address          string
	Available        bool
	BusyConnections  int32
	PendingConsumers int32
}

// Balancer is an interface to switch node from the pool
type Balancer interface {
	Get(nodes []NodeMeta) (NodeMeta, error)
}

// New Pool instance
func New(bal Balancer) *Pool {
	return &Pool{
		closeFlag: false,
		nodesLock: sync.RWMutex{},
		nodes:     make(map[string]*NodePool),
		balancer:  bal,
	}
}

// Close closes the pool and waits for completion
func (p *Pool) Close() error {
	p.nodesLock.Lock()
	defer p.nodesLock.Unlock()

	p.closeFlag = true

	// notify nodes
	for _, n := range p.nodes {
		n.NotifyClose()
	}

	// wait nodes
	for _, n := range p.nodes {
		n.Close()
	}

	// delete nodes
	for i := range p.nodes {
		delete(p.nodes, i)
	}
	return nil
}

// AddNode adds a new node to the pool
func (p *Pool) AddNode(addr string, concurrency uint16, copts conn.Options) bool {
	p.nodesLock.Lock()
	defer p.nodesLock.Unlock()

	if p.closeFlag {
		return false
	}

	if _, exists := p.nodes[addr]; !exists {
		p.nodes[addr] = newNodePool(
			addr,
			copts,
			int32(concurrency),
		)
		return true
	}
	return false
}

// Get gets free *conn.Conn from the pool
func (p *Pool) Get(timeout time.Duration) (*conn.Conn, func(), error) {
	p.nodesLock.RLock()

	// nodes meta
	meta := make([]NodeMeta, len(p.nodes))
	{
		i := 0
		for k, v := range p.nodes {
			meta[i] = NodeMeta{
				Address:          k,
				Available:        v.Available(),
				BusyConnections:  v.ConsumedConnections(),
				PendingConsumers: v.PendingConsumers(),
			}
		}
	}

	// select node
	m, err := p.balancer.Get(meta)
	if err != nil {
		p.nodesLock.RUnlock()
		return nil, nil, err
	}

	// find node by selected meta
	n, ok := p.nodes[m.Address]
	if !ok {
		p.nodesLock.RUnlock()
		return nil, nil, fmt.Errorf("failed to get node " + m.Address)
	}

	p.nodesLock.RUnlock()

	// get connection from selected node
	ret, err := n.Get(timeout)
	if err != nil {
		return nil, nil, err
	}

	return ret.Conn(), func() {
		ret.Close()
	}, nil
}
