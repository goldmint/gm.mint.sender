package types

import (
	sumuslib "github.com/void616/gm-sumuslib"
)

// PutWallet model
type PutWallet struct {
	PublicKeys []sumuslib.PublicKey
}

// ListWalletsItem model
type ListWalletsItem struct {
	PublicKey sumuslib.PublicKey
}

// DeleteWallet model
type DeleteWallet struct {
	PublicKeys []sumuslib.PublicKey
}
