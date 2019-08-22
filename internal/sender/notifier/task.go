package notifier

import (
	"time"

	"github.com/void616/gotask"
)

// Task loop
func (n *Notifier) Task(token *gotask.Token) {

	for !token.Stopped() {

		// get list
		list, err := n.dao.ListUnnotifiedSendings(itemsPerShot)
		if err != nil {
			n.logger.WithError(err).Error("Failed to get unnotified requests")
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
		for _, snd := range list {
			if out {
				break
			}

			// metrics
			t := time.Now()

			// mark as notified
			if err := n.dao.SetSendingNotified(snd.ID, true); err != nil {
				n.logger.
					WithError(err).
					WithField("id", snd.ID).
					Error("Failed to update request")
				token.Sleep(time.Second * 30)
				out = true
				continue
			}

			// notify
			notiError := "Transaction failed"
			if snd.Sent {
				notiError = ""
			}
			if !n.transporter.PublishSentEvent(snd.RequestID, snd.Sent, notiError, snd.Digest) {

				n.logger.WithField("id", snd.ID).Error("Failed to notify")

				// mark as unnotified
				if err := n.dao.SetSendingNotified(snd.ID, false); err != nil {
					n.logger.
						WithError(err).
						WithField("id", snd.ID).
						Error("Failed to update request")
					token.Sleep(time.Second * 30)
					out = true
					continue
				}
			}

			// metrics
			if n.mtxTaskDuration != nil {
				n.mtxTaskDuration.WithLabelValues("notifier_shot").Observe(time.Since(t).Seconds())
			}

			n.logger.WithField("id", snd.ID).Debug("Notified")
		}

		// metrics
		if n.mtxQueueGauge != nil {
			n.mtxQueueGauge.WithLabelValues("notifier_shot").Set(0)
		}
	}
}
