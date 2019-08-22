package nats

import (
	"time"

	proto "github.com/golang/protobuf/proto"
	walletsvc "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// NotifyRefilling sends event
func (s *Service) NotifyRefilling(addr sumuslib.PublicKey, t sumuslib.Token, a *amount.Amount, tx sumuslib.Digest) bool {
	
	// metrics
	mt := time.Now()

	reqModel := &walletsvc.RefilledEvent{
		Pubkey:      sumuslib.Pack58(addr[:]),
		Token:       t.String(),
		Amount:      a.String(),
		Transaction: sumuslib.Pack58(tx[:]),
	}

	req, err := proto.Marshal(reqModel)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal")
		return false
	}

	msg, err := s.natsConnection.Request(s.subjPrefix+walletsvc.SubjectReceived, req, time.Second*5)
	if err != nil {
		s.logger.WithError(err).Error("Failed to send request")
		return false
	}

	repModel := walletsvc.RefilledEventReply{}
	if err := proto.Unmarshal(msg.Data, &repModel); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal")
		return false
	}

	if !repModel.GetSuccess() {
		s.logger.Errorf("Unsuccessful reply: %v", repModel.GetError())
		return false
	}

	// metrics
	if s.mtxTaskDuration != nil {
		s.mtxTaskDuration.WithLabelValues("nats_notifier").Observe(time.Since(mt).Seconds())
	}

	return true
}
