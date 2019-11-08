package nats

import (
	"time"

	gonats "github.com/nats-io/go-nats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	senderNats "github.com/void616/gm-mint-sender/pkg/sender/nats"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gotask"
)

// Nats exposes endpoints via Nats server
type Nats struct {
	logger         *logrus.Entry
	subjPrefix     string
	api            API
	natsConnection *gonats.Conn
	metrics        *Metrics
}

// API provides ability to interact with service API
type API interface {
	EnqueueSendingNats(id, service string, to sumuslib.PublicKey, a *amount.Amount, t sumuslib.Token) (dup, success bool)
}

// New instance
func New(
	url string,
	subjPrefix string,
	api API,
	logger *logrus.Entry,
) (*Nats, func(), error) {

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
	logger.Infof("Nats transport enabled: %v", url)

	n := &Nats{
		logger:         logger,
		subjPrefix:     subjPrefix,
		api:            api,
		natsConnection: natsConnection,
	}
	return n, func() {
		n.natsConnection.Close()
	}, nil
}

// Metrics data
type Metrics struct {
	RequestDuration      *prometheus.HistogramVec
	NotificationDuration prometheus.Histogram
}

// AddMetrics adds metrics counters and should be called before service launch
func (n *Nats) AddMetrics(m *Metrics) {
	n.metrics = m
}

// Task loop
func (n *Nats) Task(token *gotask.Token) {

	nc := n.natsConnection

	// sub for sending requests
	_, err := nc.Subscribe(n.subjPrefix+senderNats.Send{}.Subject(), n.subSendRequest)
	if err != nil {
		n.logger.WithError(err).Errorf("Failed to subscribe to %v", n.subjPrefix+senderNats.Send{}.Subject())
	}

	// wait
	for !token.Stopped() {
		token.Sleep(time.Millisecond * 500)
	}

	// drain connection
	if err := nc.Drain(); err != nil {
		n.logger.WithError(err).Error("Failed to drain connection")
	} else {
		// wait draining
		for nc.IsDraining() {
			time.Sleep(time.Millisecond * 100)
		}
	}
}
