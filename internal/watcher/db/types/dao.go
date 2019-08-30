package types

import (
	sumuslib "github.com/void616/gm-sumuslib"
)

// WalletServices model
type WalletServices struct {
	Services  []string
	PublicKey sumuslib.PublicKey
}
