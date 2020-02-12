package conn

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
)

// receiveBytes awaits next incoming message bytes
func (c *Conn) receiveBytes(ctx context.Context) ([]byte, error) {
	ch := contextRecvChan(ctx)
	if ch == nil {
		return nil, fmt.Errorf("context has no receiving channel")
	}
	select {
	case msg, ok := <-ch:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-ctx.Done():
		return nil, context.Canceled
	case <-c.stopper:
		return nil, io.EOF
	}
}

// Receive upgrades current client context and notifies recv() to start to provide incoming messages.
// Returned channel could be closed here if connection isn't alive anymore
func (c *Conn) Receive(ctx context.Context) (rctx context.Context, cancel func()) {
	c.recvLock.Lock()
	defer c.recvLock.Unlock()
	// one receiver at one time
	if c.recvContext != nil {
		panic(errors.New("incorrect usage, concurrent receiving"))
	}

	// client context has receiving channel
	if contextRecvChan(ctx) != nil {
		panic(errors.New("incorrect usage, context already has receiving channel"))
	}

	// client should be able to cancel receiving
	rctx, rctxCancel := context.WithCancel(ctx)

	// allocate and inject receiving channel (will be closed by recv to notify client about closure)
	recvChan := make(chan []byte)
	rctx = context.WithValue(rctx, recvContextChan, recvChan)

	// new recv's client context
	c.recvContext = rctx

	return rctx, func() {
		// notify recv we don't need messages
		rctxCancel()

		c.recvLock.Lock()
		defer c.recvLock.Unlock()
		// clear recv's client context
		c.recvContext = nil
	}
}

// recv continuously receives messages
func (c *Conn) recv(reader io.Reader) error {

	// someone listens and should be notified
	defer func() {
		c.recvLock.Lock()
		if c.recvContext != nil {
			if ch := contextRecvChan(c.recvContext); ch != nil {
				close(ch)
			}
		}
		c.recvLock.Unlock()
	}()

	// message by message
	brd := bufio.NewReader(reader)
	for {
		msg, err := brd.ReadBytes(terminator)
		if err != nil {
			if err == io.EOF || c.Stopping() {
				return nil
			}
			return err
		}

		// non-empty message, someone listens, try to send message
		if len(msg) > 1 {
			msg = msg[:len(msg)-1]
			c.log(fmt.Sprintln("Recv message:", string(msg)))

			c.recvLock.Lock()
			if c.recvContext != nil {
				if ch := contextRecvChan(c.recvContext); ch != nil {
					select {
					case ch <- msg:
					case <-c.recvContext.Done():
						c.log(fmt.Sprintf("Recv context cancelled: %v", c.recvContext.Err()))
					case <-c.stopper:
					}
				}
			}
			c.recvLock.Unlock()
		}

		if c.Stopping() {
			return nil
		}
	}
}

// ---

type recvContextChanType int

var recvContextChan recvContextChanType

func contextRecvChan(ctx context.Context) chan []byte {
	v := ctx.Value(recvContextChan)
	if v == nil {
		return nil
	}
	return v.(chan []byte)
}
