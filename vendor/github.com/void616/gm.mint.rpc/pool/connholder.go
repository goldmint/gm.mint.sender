package pool

import (
	"time"

	"github.com/void616/gm.mint.rpc/conn"
)

// connHolder holds and controls underlying RPC connection
type connHolder struct {
	conn *conn.Conn
}

// newConnHolder creates connHolder instance and tries to check underlying RPC connection
func newConnHolder(addr string, opts conn.Options) (*connHolder, error) {
	ch := &connHolder{
		conn: nil,
	}
	c, err := conn.New(addr, opts)
	if err != nil {
		return nil, err
	}
	go c.Serve()
	if err := c.Heartbeat(time.Second * 10); err != nil {
		return nil, err
	}
	ch.conn = c
	return ch, nil
}

// Dead checks if underlying connection is dead
func (c *connHolder) Dead() bool {
	return c.conn.Stopping()
}

// Close closes underlying connection
func (c *connHolder) Close() {
	c.conn.Close()
}
