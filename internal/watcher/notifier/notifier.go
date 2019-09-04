package notifier

import (
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
}

// Transporter delivers notifications or fail with an error
type Transporter interface {
	NotifyRefilling(service string, to, from sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) error
}

// New Notifier instance
func New(
	dao db.DAO,
	trans Transporter,
	logger *logrus.Entry,
) (*Notifier, error) {
	n := &Notifier{
		logger:      logger,
		dao:         dao,
		transporter: trans,
	}
	return n, nil
}
