package model

import (
	"fmt"
	"math/big"
	"time"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// Sending model
type Sending struct {
	ID            uint64     `gorm:"PRIMARY_KEY;AUTO_INCREMENT:true;NOT NULL"`
	Transport     uint8      `gorm:"NOT NULL"`
	Service       string     `gorm:"SIZE:64;NOT NULL"`
	Status        uint8      `gorm:"NOT NULL"`
	To            []byte     `gorm:"SIZE:32;NOT NULL"`
	Amount        string     `gorm:"NOT NULL" sql:"TYPE:decimal(30,18)"`
	Token         uint16     `gorm:"NOT NULL"`
	Sender        []byte     `gorm:"SIZE:32"`
	SenderNonce   *uint64    `gorm:""`
	Digest        []byte     `gorm:"SIZE:32"`
	SentAtBlock   []byte     `gorm:"SIZE:32"`
	Block         []byte     `gorm:"SIZE:32"`
	RequestID     string     `gorm:"SIZE:64;NOT NULL"`
	CallbackURL   string     `gorm:"SIZE:256;NOT NULL"`
	FirstNotifyAt *time.Time `gorm:""`
	NotifyAt      *time.Time `gorm:""`
	Notified      bool       `gorm:"NOT NULL"`
}

// MapFrom mapping
func (s *Sending) MapFrom(t *types.Sending) error {
	s.ID = t.ID
	s.Transport = uint8(t.Transport)
	s.Status = uint8(t.Status)
	s.To = t.To.Bytes()
	s.Amount = t.Amount.String()
	s.Token = uint16(t.Token)
	if t.Sender != nil {
		s.Sender = (*t.Sender).Bytes()
	} else {
		s.Sender = nil
	}
	if t.SenderNonce != nil {
		s.SenderNonce = new(uint64)
		*s.SenderNonce = *t.SenderNonce
	} else {
		s.SenderNonce = nil
	}
	if t.Digest != nil {
		s.Digest = (*t.Digest).Bytes()
	} else {
		s.Digest = nil
	}
	if t.SentAtBlock != nil {
		s.SentAtBlock = t.SentAtBlock.Bytes()
	} else {
		s.SentAtBlock = nil
	}
	if t.Block != nil {
		s.Block = t.Block.Bytes()
	} else {
		s.Block = nil
	}
	s.Service = LimitStringField(t.Service, 64)
	s.RequestID = LimitStringField(t.RequestID, 64)
	s.CallbackURL = LimitStringField(t.CallbackURL, 256)
	s.FirstNotifyAt = t.FirstNotifyAt
	s.NotifyAt = t.NotifyAt
	s.Notified = t.Notified
	return nil
}

// MapTo mapping
func (s *Sending) MapTo() (*types.Sending, error) {
	var sender *sumuslib.PublicKey
	var digest *sumuslib.Digest
	var sentAtBlock *big.Int
	var block *big.Int

	to, err := sumuslib.BytesToPublicKey(s.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to")
	}

	if len(s.Sender) > 0 {
		v, err := sumuslib.BytesToPublicKey(s.Sender)
		if err != nil {
			return nil, fmt.Errorf("invalid sender")
		}
		sender = &v
	}

	if len(s.Digest) > 0 {
		v, err := sumuslib.BytesToDigest(s.Digest)
		if err != nil {
			return nil, fmt.Errorf("invalid digest")
		}
		digest = &v
	}

	if len(s.SentAtBlock) > 0 {
		sentAtBlock = new(big.Int).SetBytes(s.SentAtBlock)
	}

	if len(s.Block) > 0 {
		block = new(big.Int).SetBytes(s.Block)
	}

	amo, err := amount.FromString(s.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount")
	}

	return &types.Sending{
		ID:            s.ID,
		Transport:     types.SendingTransport(s.Transport),
		Status:        types.SendingStatus(s.Status),
		To:            to,
		Amount:        amo,
		Token:         sumuslib.Token(s.Token),
		Sender:        sender,
		SenderNonce:   s.SenderNonce,
		Digest:        digest,
		SentAtBlock:   sentAtBlock,
		Block:         block,
		Service:       s.Service,
		RequestID:     s.RequestID,
		CallbackURL:   s.CallbackURL,
		FirstNotifyAt: s.FirstNotifyAt,
		NotifyAt:      s.NotifyAt,
		Notified:      s.Notified,
	}, nil
}
