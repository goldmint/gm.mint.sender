package txconfirmer

import (
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/mint/blockparser"
	"github.com/void616/gm-mint-sender/internal/sender/db"
)

// Confirmer confirms sent transacions and updates them on DB
type Confirmer struct {
	logger *logrus.Entry
	in     <-chan *blockparser.Transaction
	dao    db.DAO
}

// New Confirmer instance
func New(
	in <-chan *blockparser.Transaction,
	dao db.DAO,
	logger *logrus.Entry,
) (*Confirmer, error) {
	f := &Confirmer{
		logger: logger,
		in:     in,
		dao:    dao,
	}
	return f, nil
}
