package model

import (
	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
	mint "github.com/void616/gm.mint"
)

// WalletSub contains data to add/remove a pair wallet:service to transaction saver
type WalletSub struct {
	PublicKey mint.PublicKey
	Service   types.Service
	Add       bool
}
