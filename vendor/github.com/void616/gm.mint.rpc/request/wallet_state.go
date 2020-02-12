package request

import (
	"context"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// WalletState model
type WalletState struct {
	Exist                 bool     `json:"exist"`
	Balance               Balance  `json:"balance"`
	Tags                  []string `json:"tags"`
	LastTransactionID     uint64   `json:"last_transaction_id"`
	LastPoolTransactionID uint64   `json:"last_pool_transaction_id"`
}

// GetWalletState method
func GetWalletState(ctx context.Context, c *conn.Conn, w mint.PublicKey) (res WalletState, rerr *rpc.Error, err error) {
	res, rerr, err = WalletState{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		PublicKey string `json:"public_key"`
	}{
		w.String(),
	}

	msg, err := c.Request(rctx, "get_wallet_state", params)
	if err != nil {
		return
	}

	switch m := msg.(type) {
	case *rpc.Error:
		rerr = m
		return
	case *rpc.Result:
		err = m.Parse(&res)
		if err == nil {
			res.Balance.checkValues()
		}
		return
	}
	return
}
