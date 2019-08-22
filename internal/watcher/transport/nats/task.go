package nats

import (
	"time"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	walletNats "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gotask"
)

// Task loop
func (s *Service) Task(token *gotask.Token) {

	nc := s.natsConnection

	// TODO: metrics, per subscription queue size

	// sub for wallet add/remove ops
	_, err := nc.Subscribe(s.subjPrefix+walletNats.SubjectWatch, func(m *gonats.Msg) {
		// metrics
		if s.mtxRequestDuration != nil {
			defer func(t time.Time, method string) {
				s.mtxRequestDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
			}(time.Now(), m.Subject)
		}

		// parse
		req := walletNats.AddRemoveRequest{}
		if err := proto.Unmarshal(m.Data, &req); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal request")
			return
		}

		s.logger.WithField("data", req.String()).Trace("Got wallet request")

		// reply
		var replyError string
		defer func() {
			rep := walletNats.AddRemoveReply{
				Success: replyError == "",
				Error:   replyError,
			}
			if b, err := proto.Marshal(&rep); err != nil {
				s.logger.WithError(err).Error("Failed to marshal reply")
			} else {
				if err := nc.Publish(m.Reply, b); err != nil {
					s.logger.WithError(err).Error("Failed to publish reply")
				}
			}
		}()

		// unpack base58
		pubs := make([]sumuslib.PublicKey, 0)
		for _, p := range req.GetPubkey() {
			b, err := sumuslib.UnpackAddress58(p)
			if err != nil {
				replyError = "One or more invalid Base58 public keys"
				return
			}
			pubs = append(pubs, b)
		}
		if req.GetAdd() {
			if !s.walletService.AddWallet(pubs...) {
				replyError = "Internal failure"
				return
			}
		} else {
			if !s.walletService.RemoveWallet(pubs...) {
				replyError = "Internal failure"
				return
			}
		}
	})
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
