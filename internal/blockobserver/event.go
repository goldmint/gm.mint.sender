package blockobserver

import (
	"encoding/json"
	"errors"
	"math/big"
)

// parseNewBlockEvent parses Sumus RPC event and returns the latest available block ID.
// `id` is nil in case of unexpected event.
func parseNewBlockEvent(msg []byte) (id *big.Int, err error) {
	id = nil
	err = nil

	model := struct {
		ID          string `json:"id,omitempty"`
		BlocksCount string `json:"block_count,omitempty"`
		LastBlockID string `json:"last_block_id,omitempty"`
	}{}
	if err = json.Unmarshal(msg, &model); err != nil {
		return
	}

	// ensure event ID
	if model.ID != "new-blocks-synchronized" {
		return
	}
	// valid block ID
	if model.LastBlockID == "" {
		err = errors.New("Last block ID is empty")
		return
	}
	// parse ID
	x, ok := big.NewInt(0).SetString(model.LastBlockID, 10)
	if !ok {
		err = errors.New("Failed to parse last block ID")
		return
	}
	id = x
	return
}
