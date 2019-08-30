package notifier

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

const itemsPerShot = 50

// Notifier sends refilling notifications
type Notifier struct {
	logger      *logrus.Entry
	transporter Transporter
	dao         db.DAO

	mtxTaskDuration *prometheus.SummaryVec
	mtxQueueGauge   *prometheus.GaugeVec
}

// Transporter delivers notifications or fail with an error
type Transporter interface {
	NotifyRefilling(service string, addr sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) error
}

// New Notifier instance
func New(
	dao db.DAO,
	trans Transporter,
	mtxTaskDuration *prometheus.SummaryVec,
	mtxQueueGauge *prometheus.GaugeVec,
	logger *logrus.Entry,
) (*Notifier, error) {
	n := &Notifier{
		logger:          logger,
		dao:             dao,
		transporter:     trans,
		mtxTaskDuration: mtxTaskDuration,
		mtxQueueGauge:   mtxQueueGauge,
	}
	return n, nil
}
