package model

import (
	"fmt"
	"math/big"
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// Incoming model
type Incoming struct {
	Base
	ID            uint64     `gorm:"PRIMARY_KEY;AUTO_INCREMENT:true;NOT NULL"`
	Service       string     `gorm:"SIZE:64;NOT NULL"`
	To            []byte     `gorm:"SIZE:32;NOT NULL"`
	From          []byte     `gorm:"SIZE:32;NOT NULL"`
	Amount        string     `gorm:"NOT NULL" sql:"TYPE:decimal(30,18)"`
	Token         uint16     `gorm:"NOT NULL"`
	Digest        []byte     `gorm:"SIZE:32;NOT NULL"`
	Block         []byte     `gorm:"SIZE:32;NOT NULL"`
	Timestamp     time.Time  `gorm:"NOT NULL"`
	FirstNotifyAt *time.Time `gorm:""`
	NotifyAt      *time.Time `gorm:""`
	Notified      bool       `gorm:"NOT NULL"`
}

// MapFrom mapping
func (i *Incoming) MapFrom(t *types.Incoming) error {
	i.ID = t.ID
	i.Service = LimitStringField(t.Service, 64)
	i.To = t.To.Bytes()
	i.From = t.From.Bytes()
	i.Amount = t.Amount.String()
	i.Token = uint16(t.Token)
	i.Digest = t.Digest.Bytes()
	i.Block = t.Block.Bytes()
	i.Timestamp = t.Timestamp
	i.FirstNotifyAt = t.FirstNotifyAt
	i.NotifyAt = t.NotifyAt
	i.Notified = t.Notified
	return nil
}

// MapTo mapping
func (i *Incoming) MapTo() (*types.Incoming, error) {
	to, err := sumuslib.BytesToPublicKey(i.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to")
	}
	from, err := sumuslib.BytesToPublicKey(i.From)
	if err != nil {
		return nil, fmt.Errorf("invalid from")
	}
	amo, err := amount.FromString(i.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount")
	}
	digest, err := sumuslib.BytesToDigest(i.Digest)
	if err != nil {
		return nil, fmt.Errorf("invalid digest")
	}
	block := new(big.Int).SetBytes(i.Block)

	return &types.Incoming{
		ID:            i.ID,
		Service:       i.Service,
		To:            to,
		From:          from,
		Amount:        amo,
		Token:         sumuslib.Token(i.Token),
		Digest:        digest,
		Block:         block,
		Timestamp:     i.Timestamp,
		FirstNotifyAt: i.FirstNotifyAt,
		NotifyAt:      i.NotifyAt,
		Notified:      i.Notified,
	}, nil
}
