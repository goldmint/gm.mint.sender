package request

import (
	"context"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// BlockchainNode model
type BlockchainNode struct {
	Index     uint32         `json:"index"`
	PublicKey mint.PublicKey `json:"public_key"`
	IP        string         `json:"ip"`
}

// GetBlockchainNodes method
func GetBlockchainNodes(ctx context.Context, c *conn.Conn) (res []BlockchainNode, rerr *rpc.Error, err error) {
	res, rerr, err = []BlockchainNode{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	msg, err := c.Request(rctx, "get_blockchain_nodes", nil)
	if err != nil {
		return
	}

	switch m := msg.(type) {
	case *rpc.Error:
		rerr = m
		return
	case *rpc.Result:
		err = m.Parse(&res)
		return
	}
	return
}
