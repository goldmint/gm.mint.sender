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
	logger      *logrus.Entry
	transporter Transporter
	dao         db.DAO
}

// Transporter delivers notifications or fail with an error
type Transporter interface {
	PublishSentEvent(
		success bool,
		err string,
		service, requestID string,
		to sumuslib.PublicKey,
		token sumuslib.Token,
		amo *amount.Amount,
		digest *sumuslib.Digest,
	) error
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
