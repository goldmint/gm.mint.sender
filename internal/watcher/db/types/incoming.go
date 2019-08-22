package types

import (
	"math/big"
	"time"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// PutIncoming model
type PutIncoming struct {
	To        sumuslib.PublicKey
	From      sumuslib.PublicKey
	Amount    *amount.Amount
	Token     sumuslib.Token
	Digest    sumuslib.Digest
	Block     *big.Int
	Timestamp time.Time
}

// MarkIncomingSent model
type MarkIncomingSent struct {
	Digest sumuslib.Digest
	Sent   bool
}

// ListUnsentIncomingsItem model
type ListUnsentIncomingsItem struct {
	To        sumuslib.PublicKey
	From      sumuslib.PublicKey
	Amount    *amount.Amount
	Token     sumuslib.Token
	Digest    sumuslib.Digest
	Block     *big.Int
	Timestamp time.Time
}
