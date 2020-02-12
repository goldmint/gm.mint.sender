package txconfirmer

import (
	"time"

	"github.com/void616/gm.mint/transaction"
	"github.com/void616/gotask"
)

// Task loop
func (c *Confirmer) Task(token *gotask.Token) {

	empty := false

	for !token.Stopped() || !empty {

		empty = false
		confirmedItems := 0
		for !empty {
			select {
			case tx := <-c.in:
				switch tx.Type {

				case transaction.TransferAssetTx:
					// save to death
					saved := false
					for !token.Stopped() && !saved {
						if err := c.dao.SetSendingConfirmed(tx.Digest, tx.From, tx.Block); err != nil {
							c.logger.WithError(err).WithField("digest", tx.Digest.String()).Errorf("Failed to confirm transaction")
							token.Sleep(time.Second * 10)
						} else {
							saved = true
							confirmedItems++
						}
					}

				case transaction.SetWalletTagTx:
					// save to death
					saved := false
					for !token.Stopped() && !saved {
						if err := c.dao.SetApprovementConfirmed(tx.Digest, tx.From, tx.Block); err != nil {
							c.logger.WithError(err).WithField("digest", tx.Digest.String()).Errorf("Failed to confirm transaction")
							token.Sleep(time.Second * 10)
						} else {
							saved = true
							confirmedItems++
						}
					}
				}
			case <-time.After(time.Second):
				empty = true
			}
		}
		if confirmedItems > 0 {
			c.logger.Infof("Confirmed %v transactions", confirmedItems)
		}
	}
}
