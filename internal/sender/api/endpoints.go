package api

import (
	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// EnqueueSendingNats adds a sending to the sender queue
func (a *API) EnqueueSendingNats(id, service string, to sumuslib.PublicKey, amo *amount.Amount, token sumuslib.Token) (dup, success bool) {
	snd := &types.Sending{
		Transport: types.SendingNats,
		Status:    types.SendingEnqueued,
		To:        to,
		Token:     token,
		Amount:    amount.FromAmount(amo),
		Service:   service,
		RequestID: id,
	}

	if err := a.dao.PutSending(snd); err != nil {
		if a.dao.DuplicateError(err) {
			return true, false
		}
		a.logger.WithError(err).Error("Failed to enqueue sending")
		return false, false
	}
	return false, true
}

// EnqueueSendingHTTP adds a sending to the sender queue
func (a *API) EnqueueSendingHTTP(id, service, callback string, to sumuslib.PublicKey, amo *amount.Amount, token sumuslib.Token) (dup, success bool) {
	snd := &types.Sending{
		Transport:   types.SendingHTTP,
		Status:      types.SendingEnqueued,
		To:          to,
		Token:       token,
		Amount:      amount.FromAmount(amo),
		Service:     service,
		RequestID:   id,
		CallbackURL: callback,
	}

	if err := a.dao.PutSending(snd); err != nil {
		if a.dao.DuplicateError(err) {
			return true, false
		}
		a.logger.WithError(err).Error("Failed to enqueue sending")
		return false, false
	}
	return false, true
}
