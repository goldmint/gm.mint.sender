package db

import (
	"math/big"

	"github.com/void616/gm-sumuslib/amount"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// DAO is a DB interface
type DAO interface {
	Available() bool
	DuplicateError(err error) bool
	MaxPacketError(err error) bool
	// SaveSenderWallet saves sending wallet to track it's outgoing transactions later
	SaveSenderWallet(sender sumuslib.PublicKey) error
	// ListSenderWallets get s list of all known senders
	ListSenderWallets() ([]sumuslib.PublicKey, error)
	// EarliestBlock finds a minimal block ID at which a transaction has been sent
	EarliestBlock() (*types.EarliestBlock, error)
	// LatestSenderNonce gets max used nonce for specified sender or zero
	LatestSenderNonce(sender sumuslib.PublicKey) (uint64, error)
	// EnqueueSending adds sending request
	EnqueueSending(request string, to sumuslib.PublicKey, amount *amount.Amount, token sumuslib.Token) error
	// ListEnqueuedSendings gets a list of enqueued sending requests
	ListEnqueuedSendings(max uint16) ([]*types.ListEnqueuedSendingsItem, error)
	// SetSendingPosted marks request as posted to the blockchain
	SetSendingPosted(id uint64, sender sumuslib.PublicKey, nonce uint64, digest sumuslib.Digest, block *big.Int) error
	// SetSendingFailed marks request as failed
	SetSendingFailed(id uint64) error
	// ListStaleSendings gets a list of stale posted requests
	ListStaleSendings(elderThanBlockID *big.Int, max uint16) ([]*types.ListStaleSendingsItem, error)
	// SetSendingConfirmed marks request as confirmed, e.g. fixed on blockchain
	SetSendingConfirmed(sender sumuslib.PublicKey, digest sumuslib.Digest, block *big.Int) error
	// ListUnnotifiedSendings gets a list of requests without notification of requestor
	ListUnnotifiedSendings(max uint16) ([]*types.ListUnnotifiedSendingsItem, error)
	// SetSendingNotified marks a sending as finally completed (requestor is notified)
	SetSendingNotified(id uint64, notified bool) error
}
