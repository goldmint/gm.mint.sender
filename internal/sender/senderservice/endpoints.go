package senderservice

import (
	"time"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// EnqueueSending adds a sending to the sender queue
func (s *Service) EnqueueSending(id string, to sumuslib.PublicKey, a *amount.Amount, t sumuslib.Token) (dup, success bool) {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "enqueue_sending")
	}

	if err := s.dao.EnqueueSending(
		id, to, a, t,
	); err != nil {
		if s.dao.DuplicateError(err) {
			return true, false
		}
		s.logger.WithError(err).Error("Failed to enqueue sending")
		return false, false
	}
	return false, true
}
