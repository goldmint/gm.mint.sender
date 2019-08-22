package txsaver

import (
	"math/big"
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gotask"
)

// Task loop
func (s *Saver) Task(token *gotask.Token) {

	empty := false
	zero := big.NewInt(0)

	for !token.Stopped() || !empty {

		empty = false
		savedItems := 0
		for !empty {
			select {
			case tx := <-s.in:
				if tx.Type != sumuslib.TransactionTransferAssets {
					break
				}

				// to
				var mTo sumuslib.PublicKey
				if tx.To != nil {
					mTo = *tx.To
				}

				// amount/token
				mToken := sumuslib.TokenGOLD
				mAmount := amount.NewAmount(tx.AmountGOLD)
				if tx.AmountMNT.Value.Cmp(zero) > 0 {
					mToken = sumuslib.TokenMNT
					mAmount.Value.Set(tx.AmountMNT.Value)
				}

				model := &types.PutIncoming{
					To:        mTo,
					From:      tx.From,
					Amount:    mAmount,
					Token:     mToken,
					Digest:    tx.Digest,
					Block:     tx.Block,
					Timestamp: tx.Timestamp,
				}

				// save to death
				saved := false
				for !token.Stopped() && !saved {
					// metrics
					t := time.Now()

					if err := s.dao.PutIncoming(model); err != nil {
						s.logger.WithError(err).WithField("digest", sumuslib.Pack58(tx.Digest[:])).Errorf("Failed to save transaction")
						token.Sleep(time.Second * 10)
					} else {
						saved = true
						savedItems++
					}

					// metrics
					if s.mtxTaskDuration != nil {
						s.mtxTaskDuration.WithLabelValues("txsaver").Observe(time.Since(t).Seconds())
					}
				}
			case <-time.After(time.Second):
				empty = true
			}
		}
		if savedItems > 0 {
			s.logger.Debugf("Saved %v transactions", savedItems)
		}
	}
}
