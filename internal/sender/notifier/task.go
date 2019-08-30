package notifier

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/sender/db/types"
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
			{
				now := time.Now().UTC()
				if snd.FirstNotifyAt == nil {
					snd.FirstNotifyAt = &now
				}
				snd.NotifyAt = &now
				snd.Notified = true
			}
			if err := n.dao.UpdateSending(snd); err != nil {
				n.logger.
					WithError(err).
					WithField("id", snd.ID).
					Error("Failed to update request")
				token.Sleep(time.Second * 30)
				out = true
				continue
			}

			notiError := "Transaction failed"
			if snd.Status == types.SendingConfirmed {
				notiError = ""
			}

			// notify
			err := n.transporter.PublishSentEvent(
				snd.Status == types.SendingConfirmed,
				notiError,
				snd.Service, snd.RequestID,
				snd.To, snd.Token, snd.Amount, snd.Digest,
			)
			if err != nil {
				n.logger.WithField("id", snd.ID).WithError(err).Error("Failed to notify")

				// notify next time
				when := time.Now().UTC()
				if snd.FirstNotifyAt != nil {
					mikes := time.Now().UTC().Sub(*snd.FirstNotifyAt).Minutes()
					switch {
					// for 5m: every 1m
					case mikes < 5:
						when = when.Add(time.Minute)
					// then for 30m: every 5m
					case mikes < 35:
						when = when.Add(time.Minute * 5)
					// then for 60m: every 10m
					case mikes < 95:
						when = when.Add(time.Minute * 10)
					// then every 120m
					default:
						when = when.Add(time.Minute * 120)
					}
				} else {
					when = when.Add(time.Hour * 24 * 365)
				}

				// mark as unnotified
				snd.NotifyAt = &when
				snd.Notified = false
				if err := n.dao.UpdateSending(snd); err != nil {
					n.logger.
						WithError(err).
						WithField("id", snd.ID).
						Error("Failed to update request")
					token.Sleep(time.Second * 30)
					out = true
					continue
				}
			} else {
				n.logger.WithField("id", snd.ID).Debug("Notified")
			}

			// metrics
			if n.mtxTaskDuration != nil {
				n.mtxTaskDuration.WithLabelValues("notifier_shot").Observe(time.Since(t).Seconds())
			}
		}

		// metrics
		if n.mtxQueueGauge != nil {
			n.mtxQueueGauge.WithLabelValues("notifier_shot").Set(0)
		}
	}
}
