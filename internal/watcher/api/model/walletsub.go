package model

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// WalletSub contains data to add/remove a pair wallet:service to transaction saver
type WalletSub struct {
	PublicKey sumuslib.PublicKey
	Service   types.Service
	Add       bool
}
