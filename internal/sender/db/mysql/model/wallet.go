package model

import (
	"fmt"

	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	mint "github.com/void616/gm.mint"
)

// Wallet model
type Wallet struct {
	PublicKey []byte `gorm:"PRIMARY_KEY;SIZE:32;NOT NULL"`
}

// MapFrom mapping
func (w *Wallet) MapFrom(t *types.Wallet) error {
	w.PublicKey = t.PublicKey.Bytes()
	return nil
}

// MapTo mapping
func (w *Wallet) MapTo() (*types.Wallet, error) {
	pub, err := mint.BytesToPublicKey(w.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &types.Wallet{
		PublicKey: pub,
	}, nil
}
