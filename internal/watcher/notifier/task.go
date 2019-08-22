package notifier

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gotask"
)

// Task loop
func (n *Notifier) Task(token *gotask.Token) {

	for !token.Stopped() {

		// get list
		list, err := n.dao.ListUnsentIncomings(itemsPerShot)
		if err != nil {
			n.logger.WithError(err).Error("Failed to get unsent items")
			token.Sleep(time.Second * 30)
			continue
		}

		// nothing
		if len(list) == 0 {
			token.Sleep(time.Second * 30)
			continue
		}

		// metrics
		if n.mtxQueueGauge != nil {
			n.mtxQueueGauge.WithLabelValues("notifier_shot").Set(float64(len(list)))
		}

		out := false
		for _, inc := range list {
			if out {
				break
			}

			// metrics
			t := time.Now()

			// mark as sent
			if err := n.dao.MarkIncomingSent(&types.MarkIncomingSent{
				Digest: inc.Digest,
				Sent:   true,
			}); err != nil {
				n.logger.
					WithError(err).
					WithField("wallet", sumuslib.Pack58(inc.To[:])).
					WithField("tx", sumuslib.Pack58(inc.Digest[:])).
					Error("Failed to update incoming")
				token.Sleep(time.Second * 30)
				out = true
				continue
			}

			// notify
			if !n.transporter.NotifyRefilling(inc.To, inc.Token, inc.Amount, inc.Digest) {

				n.logger.
					WithField("wallet", sumuslib.Pack58(inc.To[:])).
					WithField("tx", sumuslib.Pack58(inc.Digest[:])).
					Error("Failed to notify")

				// mark as unsent
				if err := n.dao.MarkIncomingSent(&types.MarkIncomingSent{
					Digest: inc.Digest,
					Sent:   false,
				}); err != nil {
					n.logger.
						WithError(err).
						WithField("wallet", sumuslib.Pack58(inc.To[:])).
						WithField("tx", sumuslib.Pack58(inc.Digest[:])).
						Error("Failed to update incoming")
					token.Sleep(time.Second * 30)
					out = true
					continue
				}
			}

			// metrics
			if n.mtxTaskDuration != nil {
				n.mtxTaskDuration.WithLabelValues("notifier_shot").Observe(time.Since(t).Seconds())
			}

			n.logger.
				WithField("wallet", sumuslib.Pack58(inc.To[:])).
				WithField("tx", sumuslib.Pack58(inc.Digest[:])).
				Info("Notified")
		}

		// metrics
		if n.mtxQueueGauge != nil {
			n.mtxQueueGauge.WithLabelValues("notifier_shot").Set(0)
		}
	}
}
