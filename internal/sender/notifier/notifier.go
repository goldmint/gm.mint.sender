package notifier

import (
	"github.com/sirupsen/logrus"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/db"
	"github.com/void616/gm.mint/amount"
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
	PublishSentEvent(ok bool, err string, service, id string, to mint.PublicKey, t mint.Token, a *amount.Amount, d *mint.Digest) error
	PublishApprovedEvent(ok bool, err string, service, id string, to mint.PublicKey, d *mint.Digest) error
}

// HTTPTransporter delivers notifications via Nats or fail with an error
type HTTPTransporter interface {
	PublishSentEvent(ok bool, err string, service, id, url string, to mint.PublicKey, t mint.Token, a *amount.Amount, d *mint.Digest) error
	PublishApprovedEvent(ok bool, err string, service, id, url string, to mint.PublicKey, d *mint.Digest) error
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
