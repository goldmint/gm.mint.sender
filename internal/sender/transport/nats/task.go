package nats

import (
	"time"
	"unicode/utf8"

	proto "github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	senderNats "github.com/void616/gm-mint-sender/pkg/sender/nats/sender"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gotask"
)

// Task loop
func (s *Service) Task(token *gotask.Token) {

	nc := s.natsConnection

	// TODO: metrics, per subscription queue size

	// sub for sending requests
	_, err := nc.Subscribe(s.subjPrefix+senderNats.SubjectSend, func(m *gonats.Msg) {
		// metrics
		if s.mtxRequestDuration != nil {
			defer func(t time.Time, method string) {
				s.mtxRequestDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
			}(time.Now(), m.Subject)
		}

		// parse
		req := senderNats.SendRequest{}
		if err := proto.Unmarshal(m.Data, &req); err != nil {
			s.logger.WithError(err).Error("Failed to unmarshal request")
			return
		}

		s.logger.WithField("data", req.String()).Trace("Got sending request")

		// reply
		var replyError string
		defer func() {
			rep := senderNats.SendReply{
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

		// check req id
		if x := utf8.RuneCountInString(req.GetId()); x < 1 || x > 64 {
			replyError = "Invalid request ID"
			return
		}

		// parse wallet address
		reqAddr, err := sumuslib.UnpackAddress58(req.GetPubkey())
		if err != nil {
			replyError = "Invalid public key"
			return
		}

		// parse token
		reqToken, err := sumuslib.ParseToken(req.GetToken())
		if err != nil {
			replyError = "Invalid token"
			return
		}

		// parse amount
		reqAmount := amount.NewFloatString(req.GetAmount())
		if reqAmount == nil {
			replyError = "Invalid amount"
			return
		}

		// enqueue
		if dups, ok := s.senderService.EnqueueSending(req.GetId(), reqAddr, reqAmount, reqToken); !ok {
			if dups {
				replyError = "Request with the same ID registered"
			} else {
				replyError = "Internal failure"
			}
			return
		}
	})
	if err != nil {
		s.logger.WithError(err).Errorf("Failed to subscribe to %v", s.subjPrefix+senderNats.SubjectSend)
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
