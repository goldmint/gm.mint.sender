package api

import (
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
)

// EnqueueSendingNats adds a sending to the sender queue
func (a *API) EnqueueSendingNats(id, service string, to mint.PublicKey, amo *amount.Amount, token mint.Token) (dup, success bool) {
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
func (a *API) EnqueueSendingHTTP(id, service, callback string, to mint.PublicKey, amo *amount.Amount, token mint.Token) (dup, success bool) {
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
