package nats

import (
	"regexp"
	"time"

	walletNats "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	"github.com/void616/gotask"
)

var serviceNameRex = regexp.MustCompile("^[a-zA-Z0-9-_]+$")

// Task loop
func (s *Service) Task(token *gotask.Token) {
	nc := s.natsConnection

	// TODO: metrics, per subscription queue size

	// sub for wallet add/remove ops
	_, err := nc.Subscribe(s.subjPrefix+walletNats.SubjectWatch, s.addRemoveWallet)
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to subscribe to %v", s.subjPrefix+walletNats.SubjectWatch)
	}

	// wait
	for !token.Stopped() {
		token.Sleep(time.Millisecond * 500)
	}

	// drain connection
	if err := nc.Drain(); err != nil {
		s.logger.WithError(err).Error("Failed to drain connection")
	} else {
		// wait draining
		for nc.IsDraining() {
			time.Sleep(time.Millisecond * 100)
		}
	}
}
