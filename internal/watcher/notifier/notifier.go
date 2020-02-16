package notifier

import (
	"github.com/sirupsen/logrus"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/watcher/db"
	"github.com/void616/gm.mint/amount"
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
	NotifyRefilling(service string, to, from mint.PublicKey, t mint.Token, a *amount.Amount, tx mint.Digest) error
}

// HTTPTransporter delivers notifications or fail with an error
type HTTPTransporter interface {
	NotifyRefilling(url, service string, to, from mint.PublicKey, t mint.Token, a *amount.Amount, tx mint.Digest) error
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
