package rpcpool

import (
	"time"

	"github.com/void616/gm-sumusrpc/conn"
	"github.com/void616/gm-sumusrpc/pool"
)

// Pool is Sumus RPC connection pool
type Pool struct {
	timeout time.Duration
	pool    *pool.Pool
}

// New Pool instance
func New(endpoints ...string) (*Pool, func(), error) {
	p := &Pool{
		timeout: time.Second * 15,
		pool:    pool.New(&pool.DefaultBalancer{}),
	}
	for _, v := range endpoints {
		p.pool.AddNode(v, 32, conn.Options{})
	}
	return p, func() {
		p.pool.Close()
	}, nil
}

// Get gets free connection from pool. *pool.Conn should be released with Close()
func (p *Pool) Get() (*pool.Conn, error) {
	c, err := p.pool.Get(p.timeout)
	if err != nil {
		return nil, err
	}
	return c, nil
}
