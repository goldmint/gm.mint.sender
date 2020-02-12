package conn

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/void616/gm.mint.rpc/rpc"
)

// ReceiveMessage receives next incoming message
func (c *Conn) ReceiveMessage(rctx context.Context) (rpc.IncomingMessage, error) {
	b, err := c.receiveBytes(rctx)
	if err != nil {
		return nil, err
	}
	return rpc.ParseMessage(b)
}

// ReceiveEvent receives next event defined by method name
func (c *Conn) ReceiveEvent(rctx context.Context, method string) (*rpc.Event, error) {
	for {
		msg, err := c.ReceiveMessage(rctx)
		if err != nil {
			return nil, err
		}
		switch m := msg.(type) {
		case *rpc.Event:
			if m.Method == method {
				return m, nil
			}
		}
	}
}

// ReceiveResponse receives next response message defined by method name and id
func (c *Conn) ReceiveResponse(rctx context.Context, method string, id uint32) (rpc.Response, error) {
	for {
		msg, err := c.ReceiveMessage(rctx)
		if err != nil {
			return nil, err
		}
		switch m := msg.(type) {
		case *rpc.Result:
			if m.Method == method && m.ID == id {
				return m, nil
			}
		case *rpc.Error:
			if m.Method == method && m.ID == id {
				return m, nil
			}
		}
	}
}

// Request sends numbered RPC request and waits for a response with exact ID, ignoring other messages
func (c *Conn) Request(rctx context.Context, method string, params interface{}) (rpc.Response, error) {
	id := atomic.LoadUint32(c.rpcRequestCounter)

	// get request bytes
	req := &rpc.Request{
		Method: method,
		ID:     id,
		Params: params,
	}
	reqBytes, err := req.JSON()
	if err != nil {
		return nil, err
	}

	// send
	if err := c.sendBytes(rctx, reqBytes); err != nil {
		return nil, err
	}
	defer atomic.AddUint32(c.rpcRequestCounter, 1)

	// wait specific method+id
	return c.ReceiveResponse(rctx, method, id)
}

// Heartbeat is request that checks connection aliveness
func (c *Conn) Heartbeat(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	msg, err := c.Request(rctx, "get_blockchain_info", nil)
	if err != nil {
		return err
	}

	switch m := msg.(type) {
	case *rpc.Result:
		if bytes.Contains(m.Result, []byte(`"blockchain_version"`)) {
			return nil
		}
		return fmt.Errorf("result fields check failed")

	case *rpc.Error:
		return m.Err()
	}

	return fmt.Errorf("not implemented")
}
