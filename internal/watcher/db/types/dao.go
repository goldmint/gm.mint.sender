package types

import (
	mint "github.com/void616/gm.mint"
)

// WalletServices model
type WalletServices struct {
	PublicKey mint.PublicKey
	Services  []Service
}
