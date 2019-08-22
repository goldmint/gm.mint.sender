package pool

import (
	"sync"
	"time"

	"github.com/void616/gm-sumusrpc/conn"
)

// Pool is a pool of pools (nodes) of connections
type Pool struct {
	closeFlag bool
	nodesLock sync.RWMutex
	nodes     map[string]*NodePool
	balancer  Balancer
}

// Balancer is an interface to switch node from the pool
type Balancer interface {
	Get(nodes map[string]*NodePool) (*NodePool, error)
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

// Get gets free connection from the pool. *Conn should be released with Close()
func (p *Pool) Get(timeout time.Duration) (*Conn, error) {

	p.nodesLock.RLock()

	// switch node
	n, err := p.balancer.Get(p.nodes)
	if err != nil {
		p.nodesLock.RUnlock()
		return nil, err
	}
	p.nodesLock.RUnlock()

	ret, err := n.Get(timeout)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
