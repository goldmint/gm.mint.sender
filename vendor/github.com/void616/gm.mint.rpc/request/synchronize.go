package request

import (
	"context"
	"strconv"

	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// Synchronize method. Mode is one of: regular, fast, manual
func Synchronize(ctx context.Context, c *conn.Conn, mode string, manualNodeIndex ...int) (rerr *rpc.Error, err error) {
	rerr, err = nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	nodes := ""
	for i, v := range manualNodeIndex {
		if i > 0 {
			nodes += ","
		}
		nodes += strconv.FormatUint(uint64(v), 10)
	}

	params := struct {
		Mode  string `json:"mode"`
		Nodes string `json:"nodes"`
	}{
		mode, nodes,
	}

	msg, err := c.Request(rctx, "synchronize", params)
	if err != nil {
		return
	}

	switch m := msg.(type) {
	case *rpc.Error:
		rerr = m
		return
	case *rpc.Result:
		return
	}
	return
}
