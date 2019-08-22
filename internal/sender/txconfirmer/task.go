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
						// metrics
						t := time.Now()

						if err := c.dao.SetSendingConfirmed(
							tx.From, tx.Digest, tx.Block,
						); err != nil {
							c.logger.WithError(err).WithField("digest", sumuslib.Pack58(tx.Digest[:])).Errorf("Failed to confirm transaction")
							token.Sleep(time.Second * 10)
						} else {
							saved = true
							confirmedItems++
						}

						// metrics
						if c.mtxTaskDuration != nil {
							c.mtxTaskDuration.WithLabelValues("txconfirmer").Observe(time.Since(t).Seconds())
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
