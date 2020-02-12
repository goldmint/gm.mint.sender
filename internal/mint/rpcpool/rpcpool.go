package rpcpool

import (
	"context"
	"time"

	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/pool"
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

// Conn returns prepared context and unused connection
func (p *Pool) Conn() (context.Context, *conn.Conn, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	con, cls, err := p.pool.Get(p.timeout)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}
	return ctx, con, func() {
		cancel()
		cls()
	}, nil
}

// ConnOnly returns only a free unused connection
func (p *Pool) ConnOnly() (*conn.Conn, func(), error) {
	con, cls, err := p.pool.Get(p.timeout)
	if err != nil {
		return nil, nil, err
	}
	return con, cls, nil
}
