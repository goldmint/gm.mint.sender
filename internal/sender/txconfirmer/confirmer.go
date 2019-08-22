package txconfirmer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/sender/db"
)

// Confirmer confirms sent transacions and updates them on DB
type Confirmer struct {
	logger *logrus.Entry
	in     <-chan *blockparser.Transaction
	dao    db.DAO

	mtxTaskDuration *prometheus.SummaryVec
}

// New Confirmer instance
func New(
	in <-chan *blockparser.Transaction,
	dao db.DAO,
	mtxTaskDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Confirmer, error) {
	f := &Confirmer{
		logger:          logger,
		in:              in,
		dao:             dao,
		mtxTaskDuration: mtxTaskDuration,
	}
	return f, nil
}
