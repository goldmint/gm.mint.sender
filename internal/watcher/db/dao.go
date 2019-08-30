package db

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
)

// DAO is a DB interface
type DAO interface {
	Available() bool
	DuplicateError(err error) bool
	MaxPacketError(err error) bool

	PutSetting(k, v string) error
	GetSetting(k, def string) (string, error)

	PutWallet(v ...*types.Wallet) error
	ListWallets() ([]*types.WalletServices, error)
	DeleteWallet(v ...*types.Wallet) error

	PutIncoming(v ...*types.Incoming) error
	ListUnnotifiedIncomings(max uint16) ([]*types.Incoming, error)
	UpdateIncoming(v *types.Incoming) error
}
