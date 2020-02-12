package request

import (
	"context"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/conn"
	"github.com/void616/gm.mint.rpc/rpc"
)

// BlockchainState model
type BlockchainState struct {
	LastBlockDigest     mint.Digest `json:"last_block_digest"`
	LastBlockMerkleRoot mint.Digest `json:"last_block_merkle_root"`
	BlockCount          *BigInt     `json:"block_count"`
	TransactionCount    *BigInt     `json:"transaction_count"`
	WalletCount         *BigInt     `json:"wallet_count"`
	NodeCount           int         `json:"node_count"`
	Balance             Balance     `json:"balance"`
	Node                struct {
		BlockchainState string `json:"blockchain_state"`
		LastError       string `json:"last_error"`
		SyncState       string `json:"sync_state"`
		ConsensusRound  int    `json:"consensus_round"`
		VotingNodes     string `json:"voting_nodes"`
		TransactionPool struct {
			PendingCount    int `json:"pending_count"`
			PendingCapacity int `json:"pending_capacity"`
			VotingCount     int `json:"voting_count"`
			VotingCapacity  int `json:"voting_capacity"`
		} `json:"transaction_pool"`
	} `json:"node"`
}

// GetBlockchainState method
func GetBlockchainState(ctx context.Context, c *conn.Conn) (res BlockchainState, rerr *rpc.Error, err error) {
	res, rerr, err = BlockchainState{}, nil, nil

	rctx, rcancel := c.Receive(ctx)
	defer rcancel()

	msg, err := c.Request(rctx, "get_blockchain_state", nil)
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
