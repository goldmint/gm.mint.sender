package walletservice

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	sumuslib "github.com/void616/gm-sumuslib"
)

// Service is a wallet service that provides methods to add/remove wallets etc.
type Service struct {
	logger       *logrus.Entry
	addWallet    chan<- sumuslib.PublicKey
	removeWallet chan<- sumuslib.PublicKey
	dao          db.DAO

	mtxMethodDuration *prometheus.SummaryVec
}

// New Service instance
func New(
	addWallet chan<- sumuslib.PublicKey,
	removeWallet chan<- sumuslib.PublicKey,
	dao db.DAO,
	mtxMethodDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Service, error) {
	f := &Service{
		logger:            logger,
		addWallet:         addWallet,
		removeWallet:      removeWallet,
		dao:               dao,
		mtxMethodDuration: mtxMethodDuration,
	}
	return f, nil
}
