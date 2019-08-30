package nats

import (
	"github.com/prometheus/client_golang/prometheus"

	gonats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// Service exposes sender service ober Nats
type Service struct {
	logger         *logrus.Entry
	subjPrefix     string
	senderService  SenderService
	natsConnection *gonats.Conn

	mtxRequestDuration *prometheus.SummaryVec
	mtxTaskDuration    *prometheus.SummaryVec
	mtxQueueGauge      *prometheus.GaugeVec
}

// SenderService provides ability to enqueue sending
type SenderService interface {
	EnqueueSending(id, service string, to sumuslib.PublicKey, a *amount.Amount, t sumuslib.Token) (dup, success bool)
}

// New Saver instance
func New(
	url string,
	subjPrefix string,
	senderService SenderService,
	mtxRequestDuration *prometheus.SummaryVec,
	mtxTaskDuration *prometheus.SummaryVec,
	mtxQueueGauge *prometheus.GaugeVec,
	logger *logrus.Entry,
) (*Service, func(), error) {

	natsConnection, err := gonats.Connect(
		url,
		gonats.Name("mint_sender"),
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
		senderService:      senderService,
		natsConnection:     natsConnection,
		mtxRequestDuration: mtxRequestDuration,
		mtxTaskDuration:    mtxTaskDuration,
		mtxQueueGauge:      mtxQueueGauge,
	}
	return f, func() {
		f.natsConnection.Close()
	}, nil
}
