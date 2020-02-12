package request

import (
	"context"

	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// BlockchainInfo model
type BlockchainInfo struct {
	BlockchainVersion     uint16   `json:"blockchain_version"`
	ClientAPIVersion      uint16   `json:"client_api_version"`
	SupportedTransactions []string `json:"supported_transactions"`
	SupportedAssets       []string `json:"supported_assets"`
	SupportedWalletTags   []string `json:"supported_wallet_tags"`
}

// GetBlockchainInfo method
func GetBlockchainInfo(ctx context.Context, c *conn.Conn) (res BlockchainInfo, rerr *rpc.Error, err error) {
	res, rerr, err = BlockchainInfo{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	msg, err := c.Request(rctx, "get_blockchain_info", nil)
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
