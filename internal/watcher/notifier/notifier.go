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
	logger    *logrus.Entry
	natsTrans NatsTransporter
	httpTrans HTTPTransporter
	dao       db.DAO
}

// NatsTransporter delivers notifications or fail with an error
type NatsTransporter interface {
	NotifyRefilling(service string, to, from sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) error
}

// HTTPTransporter delivers notifications or fail with an error
type HTTPTransporter interface {
	NotifyRefilling(url, service string, to, from sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) error
}

// New Notifier instance
func New(
	dao db.DAO,
	natsTrans NatsTransporter,
	httpTrans HTTPTransporter,
	logger *logrus.Entry,
) (*Notifier, error) {
	n := &Notifier{
		logger:    logger,
		dao:       dao,
		natsTrans: natsTrans,
		httpTrans: httpTrans,
	}
	return n, nil
}
