package model

import (
	sumuslib "github.com/void616/gm-sumuslib"
)

// WalletSub contains data to add/remove a pair wallet:service to transaction saver
type WalletSub struct {
	PublicKey sumuslib.PublicKey
	Service   string
	Add       bool
}
