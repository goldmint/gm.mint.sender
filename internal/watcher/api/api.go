package api

import (
	"github.com/sirupsen/logrus"
	"github.com/void616/gm.mint.sender/internal/watcher/api/model"
	"github.com/void616/gm.mint.sender/internal/watcher/db"
	mint "github.com/void616/gm.mint"
)

// API provides methods to interact with service
type API struct {
	logger      *logrus.Entry
	watchWallet chan<- mint.PublicKey
	walletSubs  chan<- model.WalletSub
	dao         db.DAO
}

// New instance
func New(
	watchWallet chan<- mint.PublicKey,
	walletSubs chan<- model.WalletSub,
	dao db.DAO,
	logger *logrus.Entry,
) (*API, error) {
	f := &API{
		logger:      logger,
		watchWallet: watchWallet,
		walletSubs:  walletSubs,
		dao:         dao,
	}
	return f, nil
}
