package blockobserver

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/void616/gm.mint.rpc/rpc"
)

// Parse latest blockchain block ID from node notification
func (o *Observer) parseEvent(evt *rpc.Event) (latest *big.Int, err error) {
	latest, err = nil, nil

	type Noti struct {
		Count      string `json:"count,omitempty"`
		LastID     string `json:"last_id,omitempty"`
		LastDigest string `json:"last_digest,omitempty"`
	}

	not := &Noti{}
	if err = json.Unmarshal(evt.Params, not); err != nil {
		return
	}

	if not.LastID == "" {
		err = fmt.Errorf("empty string for block number")
		return
	}

	x, ok := big.NewInt(0).SetString(not.LastID, 10)
	if !ok {
		err = fmt.Errorf("failed to parse block id: %v", not.LastID)
		return
	}

	latest = x
	return
}

// // Get node blocks count via RPC request
// func (o *Orchestrator) getNodeBlocksCount() (latest *big.Int, err error) {
// 	latest = nil
// 	err = nil

// 	ctx, conn, cls, err := o.pool.Conn()
// 	if err != nil {
// 		return
// 	}
// 	defer cls()

// 	state, rerr, err := request.GetBlockchainState(ctx, conn)
// 	if err != nil {
// 		return
// 	}
// 	if rerr != nil {
// 		err = rerr.Err()
// 		return
// 	}

// 	latest = new(big.Int).Set(state.BlockCount.Int)
// 	return
// }
