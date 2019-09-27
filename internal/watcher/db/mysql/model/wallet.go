package model

import (
	"fmt"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// Wallet model
type Wallet struct {
	Base
	PublicKey []byte  `gorm:"SIZE:32;NOT NULL"`
	ServiceID uint64  `gorm:"NOT NULL"`
	Service   Service
}

// MapFrom mapping
func (w *Wallet) MapFrom(t *types.Wallet) error {
	svc := Service{}
	if err := (&svc).MapFrom(&t.Service); err != nil {
		return err
	}

	w.PublicKey = t.PublicKey.Bytes()
	w.Service = svc
	return nil
}

// MapTo mapping
func (w *Wallet) MapTo() (*types.Wallet, error) {
	svc, err := (&w.Service).MapTo()
	if err != nil {
		return nil, err
	}
	pub, err := sumuslib.BytesToPublicKey(w.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("invalid public key")
	}
	return &types.Wallet{
		PublicKey: pub,
		Service:   *svc,
	}, nil
}
