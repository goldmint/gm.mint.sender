package txsaver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
)

// Saver saves filtered transactions to the DB
type Saver struct {
	logger          *logrus.Entry
	in              <-chan *blockparser.Transaction
	dao             db.DAO
	mtxTaskDuration *prometheus.SummaryVec
}

// New Saver instance
func New(
	in <-chan *blockparser.Transaction,
	dao db.DAO,
	mtxTaskDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Saver, error) {
	f := &Saver{
		logger:          logger,
		in:              in,
		dao:             dao,
		mtxTaskDuration: mtxTaskDuration,
	}
	return f, nil
}
