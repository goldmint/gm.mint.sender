package pool

import (
	"github.com/void616/gm-sumusrpc/conn"
)

// connHolder holds and controls underlying Sumus RPC connection
type connHolder struct {
	conn *conn.Conn
}

// newConnHolder creates connHolder instance and tries to check underlying Sumus RPC connection
func newConnHolder(addr string, opts conn.Options) (*connHolder, error) {
	ch := &connHolder{
		conn: nil,
	}
	c, err := conn.New(addr, opts)
	if err != nil {
		return nil, err
	}
	if err := c.Heartbeat(); err != nil {
		return nil, err
	}
	ch.conn = c
	return ch, nil
}

// Dead checks if underlying connection is dead
func (c *connHolder) Dead() bool {
	return c.conn.Closing()
}

// Close closes underlying connection
func (c *connHolder) Close() {
	c.conn.Close()
}
