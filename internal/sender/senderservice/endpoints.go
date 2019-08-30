package senderservice

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// EnqueueSending adds a sending to the sender queue
func (s *Service) EnqueueSending(id, service string, to sumuslib.PublicKey, a *amount.Amount, t sumuslib.Token) (dup, success bool) {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "enqueue_sending")
	}

	snd := &types.Sending{
		Status:    types.SendingEnqueued,
		To:        to,
		Token:     t,
		Amount:    amount.FromAmount(a),
		Service:   service,
		RequestID: id,
	}

	if err := s.dao.PutSending(snd); err != nil {
		if s.dao.DuplicateError(err) {
			return true, false
		}
		s.logger.WithError(err).Error("Failed to enqueue sending")
		return false, false
	}
	return false, true
}
