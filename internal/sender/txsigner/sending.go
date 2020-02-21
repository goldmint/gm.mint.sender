package txsigner

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/request"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gm.mint/fee"
	"github.com/void616/gm.mint/transaction"
)

// processSendingRequest signs and posts transaction
func (s *Signer) processSendingRequest(snd *types.Sending, currentBlock *big.Int) (posted bool) {
	posted = false

	var sigpub mint.PublicKey
	var freshNonce bool

	logger := s.logger.WithField("id", snd.ID)

	// ensure destination is approved
	if snd.Token == mint.TokenGOLD && !snd.IgnoreApprovement {
		ctx, conn, cls, err := s.pool.Conn()
		if err != nil {
			logger.WithError(err).Errorf("Failed to get free connection")
			return false
		}
		defer cls()

		ws, rerr, err := request.GetWalletState(ctx, conn, snd.To)
		if err != nil {
			logger.WithError(err).Errorf("Failed to get destination wallet state")
			return false
		}
		if rerr != nil {
			logger.WithError(rerr.Err()).Errorf("Failed to get destination wallet state")
			return false
		}

		found := false
		for _, v := range ws.Tags {
			if v == mint.WalletTagApproved.String() {
				found = true
				break
			}
		}
		if !found {
			logger.Infof("Destination is still not approved")
			return false
		}
	}

	// new tx: pick a signer
	if snd.Sender == nil {
		p, err := s.pickSendingSigner(snd.Amount, snd.Token, snd.IgnoreApprovement)
		if err != nil {
			logger.WithError(err).Errorf("Failed to pick signer")
			return false
		}
		sigpub = p
		freshNonce = true
	} else {
		// stale tx: find signer
		p := *snd.Sender
		if _, ok := s.signers[p]; !ok {
			logger.WithError(fmt.Errorf("signer %v doesn't exist", p.String())).Errorf("Failed to find signer")
			return false
		}
		sigpub = p
	}

	signer := s.signers[sigpub]
	logger = logger.WithField("signer", sigpub.StringMask())

	// new nonce or just repeat transaction
	nonce := signer.nonce + 1
	if !freshNonce {
		nonce = *snd.SenderNonce
	}
	logger = logger.WithField("nonce", nonce)

	// sign
	tatx := transaction.TransferAsset{
		Address: snd.To,
		Token:   snd.Token,
		Amount:  snd.Amount,
	}
	stx, err := tatx.Sign(signer.signer, nonce)
	if err != nil {
		logger.WithError(err).Errorf("Failed to sign transaction")
		return false
	}

	// increment signer's nonce
	if freshNonce {
		signer.nonce++

		// reduce balance
		if !signer.emitter {
			defer func() {
				if posted {
					sub := amount.FromAmount(snd.Amount)
					switch snd.Token {
					case mint.TokenGOLD:
						sub.Value.Add(sub.Value, fee.GoldFee(sub, signer.mnt).Value)
						signer.gold.Value.Sub(signer.gold.Value, sub.Value)
					case mint.TokenMNT:
						sub.Value.Add(sub.Value, fee.MntFee(sub).Value)
						signer.mnt.Value.Sub(signer.mnt.Value, sub.Value)
					}

					// metrics
					if s.metrics != nil {
						s.metrics.Balance.WithLabelValues(signer.public.String(), "gold").Set(signer.gold.Float64())
						s.metrics.Balance.WithLabelValues(signer.public.String(), "mnt").Set(signer.mnt.Float64())
					}
				}
			}()
		}
	}

	// get free connection
	ctx, conn, cls, err := s.pool.Conn()
	if err != nil {
		logger.WithError(err).Errorf("Failed to get free RPC connection")
		return false
	}
	defer cls()

	// save as posted
	snd.Status = types.SendingPosted
	snd.Sender = &mint.PublicKey{}
	*snd.Sender = signer.public
	snd.SenderNonce = new(uint64)
	*snd.SenderNonce = nonce
	snd.Digest = &mint.Digest{}
	*snd.Digest = stx.Digest
	snd.SentAtBlock = new(big.Int).Set(currentBlock)
	if err := s.dao.UpdateSending(snd); err != nil {
		logger.WithError(err).Errorf("Failed to mark request posted")
		return false
	}

	// mark as failed in some cases
	reject := false
	defer func() {
		if reject {
			snd.Status = types.SendingFailed
			if err := s.dao.UpdateSending(snd); err != nil {
				logger.WithError(err).Errorf("Failed to mark request failed")
			}
		}
	}()

	logger.Debugf("Sending transaction")

	// post
	_, rerr, err := request.AddTransaction(ctx, conn, transaction.TransferAssetTx, stx.Data)
	if err != nil {
		logger.WithError(err).Errorf("Sending failed")
		// don't reject, probably tx is posted
		return false
	}

	if rerr != nil {
		ncode, _, ok := rerr.GetReason()
		if !ok {
			logger.WithError(err).Errorf("Sending failed")
			// don't reject, probably tx is posted
			return false
		}

		switch {
		case ncode.TxAddedAlready():
			// just ok
		case ncode.TxWalletNotReady():
			logger.Errorf("Node replied with: wallet not ready")
			// fresh or repeated tx, doesn't matter, it's failed
			reject = true
			return false
		case ncode.TxNonceAhead():
			logger.Errorf("Node replied with: nonce ahead")
			// not matter, keep posting it
		case ncode.TxNonceBehind():
			logger.Errorf("Node replied with: nonce behind (duplicate)")
			// reject it in case it's a fresh tx
			if freshNonce {
				reject = true
			}
			return false
		}
	}

	signer.signedCount++
	posted = true
	return
}

// pickSendingSigner picks appropriate signer
func (s *Signer) pickSendingSigner(a *amount.Amount, t mint.Token, emitterRequired bool) (mint.PublicKey, error) {

	// all signers
	sorted := make([]mint.PublicKey, 0)
	for _, v := range s.signers {
		sorted = append(sorted, v.public)
	}

	// sort by number of signed requests (asc)
	sort.Slice(sorted, func(i, j int) bool {
		s1 := s.signers[sorted[i]]
		s2 := s.signers[sorted[j]]
		return s1.signedCount < s2.signedCount
	})

	for _, pub := range sorted {
		v := s.signers[pub]

		// emitter required
		if emitterRequired && !v.emitter {
			continue
		}

		// emitter has priority (no need to check balance)
		if v.emitter {
			return v.public, nil
		}

		send := amount.FromAmount(a)
		switch t {
		case mint.TokenGOLD:
			send.Value.Add(send.Value, fee.GoldFee(send, v.mnt).Value)
			if v.gold.Value.Cmp(send.Value) >= 0 {
				return v.public, nil
			}
		case mint.TokenMNT:
			send.Value.Add(send.Value, fee.MntFee(send).Value)
			if v.mnt.Value.Cmp(send.Value) >= 0 {
				return v.public, nil
			}
		}
	}

	return mint.PublicKey{}, errors.New("all failed or not enough token")
}
