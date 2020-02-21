package types

import (
	"math/big"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
)

// Sending model
type Sending struct {
	ID                uint64
	Transport         SendingTransport
	Status            SendingStatus
	To                mint.PublicKey
	Token             mint.Token
	Amount            *amount.Amount
	IgnoreApprovement bool
	Sender            *mint.PublicKey
	SenderNonce       *uint64
	Digest            *mint.Digest
	SentAtBlock       *big.Int
	Block             *big.Int
	Service           string
	RequestID         string
	CallbackURL       string
	FirstNotifyAt     *time.Time
	NotifyAt          *time.Time
	Notified          bool
}
