package txsigner

import (
	"fmt"
	"math/big"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.rpc/request"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gm.mint/transaction"
)

// processApprovingRequest signs and posts transaction
func (s *Signer) processApprovingRequest(apv *types.Approvement, currentBlock *big.Int) (posted bool) {
	posted = false

	var sigpub mint.PublicKey
	var freshNonce bool

	logger := s.logger.WithField("id", apv.ID)

	// new tx: pick a signer
	if apv.Sender == nil {
		p, err := s.pickApprovementSigner()
		if err != nil {
			logger.WithError(err).Errorf("Failed to pick signer")
			return false
		}
		sigpub = p
		freshNonce = true
	} else {
		// stale tx: find signer
		p := *apv.Sender
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
		nonce = *apv.SenderNonce
	}
	logger = logger.WithField("nonce", nonce)

	// sign
	tatx := transaction.SetWalletTag{
		Address: apv.To,
		Tag:     mint.WalletTagApproved,
	}
	stx, err := tatx.Sign(signer.signer, nonce)
	if err != nil {
		logger.WithError(err).Errorf("Failed to sign transaction")
		return false
	}

	// increment signer's nonce
	if freshNonce {
		signer.nonce++
	}

	// get free connection
	ctx, conn, cls, err := s.pool.Conn()
	if err != nil {
		logger.WithError(err).Errorf("Failed to get free RPC connection")
		return false
	}
	defer cls()

	// save as posted
	apv.Status = types.SendingPosted
	apv.Sender = &mint.PublicKey{}
	*apv.Sender = signer.public
	apv.SenderNonce = new(uint64)
	*apv.SenderNonce = nonce
	apv.Digest = &mint.Digest{}
	*apv.Digest = stx.Digest
	apv.SentAtBlock = new(big.Int).Set(currentBlock)
	if err := s.dao.UpdateApprovement(apv); err != nil {
		logger.WithError(err).Errorf("Failed to mark request posted")
		return false
	}

	// mark as failed in some cases
	reject := false
	defer func() {
		if reject {
			apv.Status = types.SendingFailed
			if err := s.dao.UpdateApprovement(apv); err != nil {
				logger.WithError(err).Errorf("Failed to mark request failed")
			}
		}
	}()

	logger.Debugf("Sending transaction")

	// post
	_, rerr, err := request.AddTransaction(ctx, conn, transaction.SetWalletTagTx, stx.Data)
	if err != nil {
		logger.WithError(err).Errorf("Sending failed")
		// don't reject, probably tx is posted
		return false
	}

	if rerr != nil {
		logger.WithError(rerr.Err()).Errorf("Sending failed")
		reject = true
		return false
	}

	signer.signedCount++
	posted = true
	return
}

// pickApprovementSigner picks appropriate signer
func (s *Signer) pickApprovementSigner() (mint.PublicKey, error) {

	sorted := make([]mint.PublicKey, 0)
	for _, v := range s.signers {
		if v.approver {
			sorted = append(sorted, v.public)
		}
	}

	if len(sorted) == 0 {
		return mint.PublicKey{}, fmt.Errorf("failed to find 'authority' signer")
	}

	return sorted[0], nil
}
