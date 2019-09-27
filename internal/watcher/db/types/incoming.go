package types

import (
	"math/big"
	"time"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// Incoming model
type Incoming struct {
	ID            uint64
	Service       Service
	To            sumuslib.PublicKey
	From          sumuslib.PublicKey
	Amount        *amount.Amount
	Token         sumuslib.Token
	Digest        sumuslib.Digest
	Block         *big.Int
	Timestamp     time.Time
	FirstNotifyAt *time.Time
	NotifyAt      *time.Time
	Notified      bool
}
