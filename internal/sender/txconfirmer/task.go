package txconfirmer

import (
	"time"

	sumuslib "github.com/void616/gm-sumuslib"
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
				if tx.Type == sumuslib.TransactionTransferAssets {

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
