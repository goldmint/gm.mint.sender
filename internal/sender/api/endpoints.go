package api

import (
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/request"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gm.mint/amount"
)

// EnqueueSending adds a sending to the sender queue
func (a *API) EnqueueSending(trans types.SendingTransport, id, service, callbackURL string, to mint.PublicKey, amo *amount.Amount, token mint.Token) (dup, success bool) {
	snd := &types.Sending{
		Transport:   trans,
		Status:      types.SendingEnqueued,
		To:          to,
		Token:       token,
		Amount:      amount.FromAmount(amo),
		Service:     service,
		RequestID:   id,
		CallbackURL: callbackURL,
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

// EnqueueApprovement adds an approvement to the sender queue
func (a *API) EnqueueApprovement(trans types.SendingTransport, id, service, callbackURL string, to mint.PublicKey) (dup, success bool) {

	// ensure destination is not approved yet, return success otherwise
	{
		ctx, conn, cls, err := a.pool.Conn()
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to get free connection")
			return false, false
		}
		defer cls()

		ws, rerr, err := request.GetWalletState(ctx, conn, to)
		if err != nil {
			a.logger.WithError(err).Errorf("Failed to get approving wallet state")
			return false, false
		}
		if rerr != nil {
			a.logger.WithError(rerr.Err()).Errorf("Failed to get approving wallet state")
			return false, false
		}

		for _, v := range ws.Tags {
			if v == mint.WalletTagApproved.String() {
				a.logger.Infof("Approving wallet is already approved, skipping")
				return false, true
			}
		}
	}

	apv := &types.Approvement{
		Transport:   trans,
		Status:      types.SendingEnqueued,
		To:          to,
		Service:     service,
		RequestID:   id,
		CallbackURL: callbackURL,
	}

	if err := a.dao.PutApprovement(apv); err != nil {
		if a.dao.DuplicateError(err) {
			return true, false
		}
		a.logger.WithError(err).Error("Failed to enqueue approvement")
		return false, false
	}
	return false, true
}
