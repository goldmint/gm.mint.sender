package types

import (
	"math/big"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
)

// Incoming model
type Incoming struct {
	ID            uint64
	Service       Service
	To            mint.PublicKey
	From          mint.PublicKey
	Amount        *amount.Amount
	Token         mint.Token
	Digest        mint.Digest
	Block         *big.Int
	Timestamp     time.Time
	FirstNotifyAt *time.Time
	NotifyAt      *time.Time
	Notified      bool
}
