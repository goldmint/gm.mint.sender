package nats

import (
	"time"

	gonats "github.com/nats-io/go-nats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	walletNats "github.com/void616/gm-mint-sender/pkg/watcher/nats"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gotask"
)

// Nats is Nats server transport for service API
type Nats struct {
	logger         *logrus.Entry
	subjPrefix     string
	api            API
	natsConnection *gonats.Conn
	metrics        *Metrics
}

// API provides API methods
type API interface {
	AddWallet(service string, serviceTrans types.ServiceTransport, serviceCallbackURL string, p ...sumuslib.PublicKey) bool
	RemoveWallet(service string, p ...sumuslib.PublicKey) bool
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
	logger.Infof("Nats transport enabled: %v", url)

	f := &Nats{
		logger:         logger,
		subjPrefix:     subjPrefix,
		api:            api,
		natsConnection: natsConnection,
	}
	return f, func() {
		f.natsConnection.Close()
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

	// sub for wallet add/remove ops
	subj := n.subjPrefix + walletNats.AddRemove{}.Subject()
	_, err := nc.Subscribe(subj, n.subAddRemoveWallet)
	if err != nil {
		n.logger.WithError(err).Errorf("Failed to subscribe to %v", subj)
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
