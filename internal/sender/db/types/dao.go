package types

import (
	"math/big"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// EarliestBlock model
type EarliestBlock struct {
	Block *big.Int
	Empty bool
}

// ListEnqueuedSendingsItem model
type ListEnqueuedSendingsItem struct {
	ID     uint64
	To     sumuslib.PublicKey
	Amount *amount.Amount
	Token  sumuslib.Token
}

// ListStaleSendingsItem model
type ListStaleSendingsItem struct {
	ID     uint64
	To     sumuslib.PublicKey
	Amount *amount.Amount
	Token  sumuslib.Token
	From   sumuslib.PublicKey
	Nonce  uint64
}

// ListUnnotifiedSendingsItem model
type ListUnnotifiedSendingsItem struct {
	ID        uint64
	Digest    sumuslib.Digest
	RequestID string
	Sent      bool
}
