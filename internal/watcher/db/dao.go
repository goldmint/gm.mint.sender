package db

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
)

// DAO is a DB interface
type DAO interface {
	Available() bool
	DuplicateError(err error) bool
	MaxPacketError(err error) bool
	PutWallet(*types.PutWallet) error
	ListWallets() ([]*types.ListWalletsItem, error)
	DeleteWallet(*types.DeleteWallet) error
	PutIncoming(*types.PutIncoming) error
	MarkIncomingSent(*types.MarkIncomingSent) error
	ListUnsentIncomings(max uint16) ([]*types.ListUnsentIncomingsItem, error)
}
