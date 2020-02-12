package request

import (
	"context"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// WalletsDump model
type WalletsDump struct {
	Wallets []struct {
		PublicKey mint.PublicKey `json:"k"`
		Balance   Balance        `json:"b"`
		Tags      []string       `json:"t"`
	} `json:"wallets,omitempty"`
	LocalFile string `json:"local_file,omitempty"`
}

// DumpWallets method
func DumpWallets(ctx context.Context, c *conn.Conn, intoLocalFile bool) (res WalletsDump, rerr *rpc.Error, err error) {
	res, rerr, err = WalletsDump{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	params := struct {
		IntoFile bool `json:"into_file"`
	}{
		intoLocalFile,
	}

	msg, err := c.Request(rctx, "dump_wallets", params)
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
			for _, v := range res.Wallets {
				v.Balance.checkValues()
			}
		}
		return
	}
	return
}
