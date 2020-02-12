package request

import (
	"context"
	"encoding/hex"

	"github.com/void616/gm.mint/transaction"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// AddedTransaction model
type AddedTransaction struct {
	VotingNow       bool `json:"voting_now"`
	TransactionPool struct {
		PendingCapacity int `json:"pending_capacity"`
		VotingCapacity  int `json:"voting_capacity"`
	} `json:"txpool"`
}

// AddTransaction method
func AddTransaction(ctx context.Context, c *conn.Conn, tc transaction.Code, data []byte) (res AddedTransaction, rerr *rpc.Error, err error) {
	res, rerr, err = AddedTransaction{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}{
		tc.String(),
		hex.EncodeToString(data),
	}

	msg, err := c.Request(rctx, "add_transaction", params)
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
