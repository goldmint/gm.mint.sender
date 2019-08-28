package nats

import (
	"github.com/prometheus/client_golang/prometheus"

	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	sumuslib "github.com/void616/gm-sumuslib"
)

// Service saves filtered transactions to the DB
type Service struct {
	logger         *logrus.Entry
	subjPrefix     string
	walletService  WalletService
	natsConnection *gonats.Conn

	mtxRequestDuration *prometheus.SummaryVec
	mtxTaskDuration    *prometheus.SummaryVec
	mtxQueueGauge      *prometheus.GaugeVec
}

// WalletService provides wallets management methods
type WalletService interface {
	AddWallet(...sumuslib.PublicKey) bool
	RemoveWallet(...sumuslib.PublicKey) bool
}

// New Saver instance
func New(
	url string,
	subjPrefix string,
	walletService WalletService,
	mtxRequestDuration *prometheus.SummaryVec,
	mtxTaskDuration *prometheus.SummaryVec,
	mtxQueueGauge *prometheus.GaugeVec,
	logger *logrus.Entry,
) (*Service, func(), error) {
	natsConnection, err := gonats.Connect(
		url,
		gonats.Name("mint_watcher"),
		gonats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, nil, err
	}
	natsConnection.SetReconnectHandler(func(_ *gonats.Conn) {
		logger.Warnf("Nats reconnected: %v", url)
	})
	natsConnection.SetDisconnectHandler(func(_ *gonats.Conn) {
		logger.Warnf("Nats disconnected: %v", url)
	})
	logger.Infof("Nats connected: %v", url)

	f := &Service{
		logger:             logger,
		subjPrefix:         subjPrefix,
		walletService:      walletService,
		natsConnection:     natsConnection,
		mtxRequestDuration: mtxRequestDuration,
		mtxTaskDuration:    mtxTaskDuration,
		mtxQueueGauge:      mtxQueueGauge,
	}
	return f, func() {
		f.natsConnection.Close()
	}, nil
}
