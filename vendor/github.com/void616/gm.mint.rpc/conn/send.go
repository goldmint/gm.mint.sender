package conn

import (
	"context"
	"fmt"
	"io"
	"time"
)

// sendBytes sends some message bytes
func (c *Conn) sendBytes(ctx context.Context, msg []byte) (err error) {
	defer func() {
		if err != nil {
			c.log(fmt.Sprintf("Send error: %v", err))
			c.markStopping()
			c.conn.Close()
		}
	}()

	// default writing deadline
	deadline := time.Now().Add(time.Second * 10)

	// context has own deadline
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}

	// send
	c.conn.SetWriteDeadline(deadline)
	return c.send(c.conn, msg)
}

// send sends a message
func (c *Conn) send(writer io.Writer, b []byte) error {
	// data
	if len(b) > 0 {
		n, err := writer.Write(b)
		if err != nil {
			return err
		}
		if n != len(b) {
			return fmt.Errorf("wrote %v, expected %v bytes of the message", n, len(b))
		}
		c.log(fmt.Sprintln("Send message:", string(b)))
	}
	// terminator
	var terminator = []byte{terminator}
	n, err := writer.Write(terminator)
	if err != nil {
		return err
	}
	if n != len(terminator) {
		return fmt.Errorf("wrote %v, expected %v bytes of the terminator", n, len(terminator))
	}
	return nil
}
