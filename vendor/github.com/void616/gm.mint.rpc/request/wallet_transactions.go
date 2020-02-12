package request

import (
	"context"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// BinaryWalletTransaction model
type BinaryWalletTransaction struct {
	Data  ByteArray `json:"data"`
	Block *BigInt   `json:"block"`
}

// TextualWalletTransaction model
type TextualWalletTransaction struct {
	Desc   string      `json:"desc"`
	Digest mint.Digest `json:"digest"`
	Block  *BigInt     `json:"block"`
}

// GetWalletTransactionsBinary method
func GetWalletTransactionsBinary(ctx context.Context, c *conn.Conn, w mint.PublicKey, max uint32, poolLookup, incoming, outgoing bool) (res []BinaryWalletTransaction, rerr *rpc.Error, err error) {
	res, rerr, err = []BinaryWalletTransaction{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		Binary     bool   `json:"binary"`
		PublicKey  string `json:"public_key"`
		Count      uint32 `json:"count"`
		PoolLookup bool   `json:"pool_lookup"`
		Incoming   bool   `json:"incoming"`
		Outgoing   bool   `json:"outgoing"`
	}{
		true,
		w.String(),
		max,
		poolLookup,
		incoming,
		outgoing,
	}

	msg, err := c.Request(rctx, "get_wallet_transactions", params)
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

// GetWalletTransactionsTextual method
func GetWalletTransactionsTextual(ctx context.Context, c *conn.Conn, w mint.PublicKey, max uint32, poolLookup, incoming, outgoing bool) (res []TextualWalletTransaction, rerr *rpc.Error, err error) {
	res, rerr, err = []TextualWalletTransaction{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		Binary     bool   `json:"binary"`
		PublicKey  string `json:"public_key"`
		Count      uint32 `json:"count"`
		PoolLookup bool   `json:"pool_lookup"`
		Incoming   bool   `json:"incoming"`
		Outgoing   bool   `json:"outgoing"`
	}{
		false,
		w.String(),
		max,
		poolLookup,
		incoming,
		outgoing,
	}

	msg, err := c.Request(rctx, "get_wallet_transactions", params)
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
