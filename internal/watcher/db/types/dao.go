package types

import (
	sumuslib "github.com/void616/gm-sumuslib"
)

// WalletServices model
type WalletServices struct {
	PublicKey sumuslib.PublicKey
	Services  []Service
}
