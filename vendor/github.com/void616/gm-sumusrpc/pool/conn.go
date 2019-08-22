package pool

import "github.com/void616/gm-sumusrpc/conn"

// Conn represents connection within pool
type Conn struct {
	backChan chan<- *Conn
	holder   *connHolder
}

// newConn creates new pool Conn instance
func newConn(holder *connHolder, backChan chan<- *Conn) *Conn {
	ret := &Conn{
		backChan: backChan,
		holder:   holder,
	}
	return ret
}

// Close releases connection back to the node pool
func (c *Conn) Close() error {
	c.backChan <- c
	return nil
}

// Conn gets underlying Sumus RPC connection
func (c *Conn) Conn() *conn.Conn {
	return c.holder.conn
}
