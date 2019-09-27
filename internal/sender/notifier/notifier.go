package notifier

import (
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/sender/db"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

const itemsPerShot = 50

// Notifier sends refilling notifications
type Notifier struct {
	logger          *logrus.Entry
	natsTransporter NatsTransporter
	httpTransporter HTTPTransporter
	dao             db.DAO
}

// NatsTransporter delivers notifications via Nats or fail with an error
type NatsTransporter interface {
	PublishSentEvent(ok bool, err string, service, id string, to sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, d *sumuslib.Digest) error
}

// HTTPTransporter delivers notifications via Nats or fail with an error
type HTTPTransporter interface {
	PublishSentEvent(ok bool, err string, service, id, url string, to sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, d *sumuslib.Digest) error
}

// New Notifier instance
func New(
	dao db.DAO,
	natsTrans NatsTransporter, httpTrans HTTPTransporter,
	logger *logrus.Entry,
) (*Notifier, error) {
	n := &Notifier{
		logger:          logger,
		dao:             dao,
		natsTransporter: natsTrans,
		httpTransporter: httpTrans,
	}
	return n, nil
}
