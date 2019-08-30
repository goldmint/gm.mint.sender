package txsigner

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gm-sumuslib/fee"
	"github.com/void616/gm-sumuslib/transaction"
	"github.com/void616/gm-sumusrpc/rpc"
)

// processRequest signs and posts transaction
func (s *Signer) processRequest(snd *types.Sending, currentBlock *big.Int) (posted bool) {
	posted = false

	var sigpub sumuslib.PublicKey
	var freshNonce bool

	logger := s.logger.WithField("id", snd.ID)

	// new tx: pick a signer
	if snd.Sender == nil {
		p, err := s.pickSigner(snd.Amount, snd.Token)
		if err != nil {
			logger.WithError(err).Errorln("Failed to pick signer")
			return false
		}
		sigpub = p
		freshNonce = true
	} else {
		// stale tx: find signer
		p := *snd.Sender
		if _, ok := s.signers[p]; !ok {
			logger.WithError(fmt.Errorf("signer %v doesn't exist", p.String())).Errorln("Failed to find signer")
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
	stx, err := tatx.Construct(signer.signer, nonce)
	if err != nil {
		logger.WithError(err).Errorln("Failed to sign transaction")
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
					case sumuslib.TokenGOLD:
						sub.Value.Add(sub.Value, fee.GoldFee(sub, signer.mnt).Value)
						signer.gold.Value.Sub(signer.gold.Value, sub.Value)
					case sumuslib.TokenMNT:
						sub.Value.Add(sub.Value, fee.MntFee(sub).Value)
						signer.mnt.Value.Sub(signer.mnt.Value, sub.Value)
					}

					// metrics
					if s.mtxBalanceGauge != nil {
						s.mtxBalanceGauge.WithLabelValues(signer.public.String(), "gold").Set(signer.gold.Float64())
						s.mtxBalanceGauge.WithLabelValues(signer.public.String(), "mnt").Set(signer.mnt.Float64())
					}
				}
			}()
		}
	}

	// get free connection
	conn, err := s.pool.Get()
	if err != nil {
		logger.WithError(err).Errorln("Failed to get free RPC connection")
		return false
	}
	defer conn.Close()

	// save as posted
	snd.Status = types.SendingPosted
	snd.Sender = &sumuslib.PublicKey{}
	*snd.Sender = signer.public
	snd.SenderNonce = new(uint64)
	*snd.SenderNonce = nonce
	snd.Digest = &sumuslib.Digest{}
	*snd.Digest = stx.Digest
	snd.SentAtBlock = new(big.Int).Set(currentBlock)
	if err := s.dao.UpdateSending(snd); err != nil {
		logger.WithError(err).Errorln("Failed to mark request posted")
		return false
	}

	// mark as failed in some cases
	reject := false
	defer func() {
		if reject {
			snd.Status = types.SendingFailed
			if err := s.dao.UpdateSending(snd); err != nil {
				logger.WithError(err).Errorln("Failed to mark request failed")
			}
		}
	}()

	logger.Debugln("Sending transaction")

	// post
	_, code, err := rpc.AddTransaction(conn.Conn(), sumuslib.TransactionTransferAssets, hex.EncodeToString(stx.Data))
	if err != nil {
		logger.WithError(err).Errorln("Sending failed")
		// don't reject, probably tx is posted
		return false
	}

	logger.WithField("node_code", fmt.Sprint(uint16(code)))

	if code != rpc.ECSuccess {
		atec := rpc.AddTransactionErrorCode(code)
		switch {
		case atec.AddedAlready():
			// just ok
		case atec.WalletInconsistency():
			logger.Errorln("Node replied with: wallet not ready")
			// fresh or repeated tx, doesn't matter, it's failed
			reject = true
			return false
		case atec.NonceAhead():
			logger.Errorln("Node replied with: nonce ahead")
			// not matter, keep posting it
		case atec.NonceBehind():
			logger.Errorln("Node replied with: nonce behind (duplicate)")
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

// pickSigner picks appropriate signer
func (s *Signer) pickSigner(a *amount.Amount, t sumuslib.Token) (sumuslib.PublicKey, error) {

	sorted := make([]sumuslib.PublicKey, 0)
	for _, v := range s.signers {
		sorted = append(sorted, v.public)
	}

	sort.Slice(sorted, func(i, j int) bool {
		s1 := s.signers[sorted[i]]
		s2 := s.signers[sorted[j]]
		return s1.signedCount < s2.signedCount
	})

	for _, pub := range sorted {
		v := s.signers[pub]

		if v.failed {
			continue
		}
		if v.emitter {
			return v.public, nil
		}

		send := amount.FromAmount(a)
		switch t {
		case sumuslib.TokenGOLD:
			send.Value.Add(send.Value, fee.GoldFee(send, v.mnt).Value)
			if v.gold.Value.Cmp(send.Value) >= 0 {
				return v.public, nil
			}
		case sumuslib.TokenMNT:
			send.Value.Add(send.Value, fee.MntFee(send).Value)
			if v.mnt.Value.Cmp(send.Value) >= 0 {
				return v.public, nil
			}
		}
	}

	return sumuslib.PublicKey{}, errors.New("all failed or not enough token")
}
