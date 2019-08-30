package walletservice

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	sumuslib "github.com/void616/gm-sumuslib"
)

// WalletSub contains data to add/remove a pair wallet:service to transaction saver
type WalletSub struct {
	PublicKey sumuslib.PublicKey
	Service   string
	Add       bool
}

// Service is a wallet service that provides methods to add/remove wallets etc.
type Service struct {
	logger      *logrus.Entry
	watchWallet chan<- sumuslib.PublicKey
	walletSubs  chan<- WalletSub
	dao         db.DAO

	mtxMethodDuration *prometheus.SummaryVec
}

// New Service instance
func New(
	watchWallet chan<- sumuslib.PublicKey,
	walletSubs chan<- WalletSub,
	dao db.DAO,
	mtxMethodDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Service, error) {
	f := &Service{
		logger:            logger,
		watchWallet:       watchWallet,
		walletSubs:        walletSubs,
		dao:               dao,
		mtxMethodDuration: mtxMethodDuration,
	}
	return f, nil
}
