package notifier

import (
	"sync"
	"time"

	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gotask"
)

// Task loop
func (n *Notifier) Task(token *gotask.Token) {

	var wg sync.WaitGroup
	var stopper = make(chan struct{})
	var sleep = func(d time.Duration) {
		select {
		case <-stopper:
		case <-time.After(d):
		}
	}

	// approvements
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger := n.logger.WithField("approvements", "")

		for {
			select {
			case <-stopper:
				return
			default:
			}

			// get list
			list, err := n.dao.ListUnnotifiedApprovements(itemsPerShot)
			if err != nil {
				logger.WithError(err).Error("Failed to get unnotified requests")
				sleep(time.Second * 30)
				continue
			}

			// nothing
			if len(list) == 0 {
				sleep(time.Second * 30)
				continue
			}

			out := false
			for _, snd := range list {
				if out {
					break
				}

				// mark as notified
				{
					now := time.Now().UTC()
					if snd.FirstNotifyAt == nil {
						snd.FirstNotifyAt = &now
					}
					snd.NotifyAt = &now
					snd.Notified = true
				}
				if err := n.dao.UpdateApprovement(snd); err != nil {
					logger.
						WithError(err).
						WithField("id", snd.ID).
						Error("Failed to update request")
					sleep(time.Second * 30)
					out = true
					continue
				}

				notiErrorDesc := "Transaction failed"
				if snd.Status == types.SendingConfirmed {
					notiErrorDesc = ""
				}

				// notify
				var notiErr error
				switch snd.Transport {
				case types.SendingNats:
					if n.natsTransporter != nil {
						notiErr = n.natsTransporter.PublishApprovedEvent(
							snd.Status == types.SendingConfirmed,
							notiErrorDesc,
							snd.Service, snd.RequestID,
							snd.To, snd.Digest,
						)
					} else {
						logger.Warn("Nats transport is disabled, skipping notification")
						continue
					}
				case types.SendingHTTP:
					if n.httpTransporter != nil {
						if snd.CallbackURL != "" {
							notiErr = n.httpTransporter.PublishApprovedEvent(
								snd.Status == types.SendingConfirmed,
								notiErrorDesc,
								snd.Service, snd.RequestID, snd.CallbackURL,
								snd.To, snd.Digest,
							)
						}
					} else {
						logger.Warn("HTTP transport is disabled, skipping notification")
						continue
					}
				default:
					logger.Errorf("Transport %v is not implemented", snd.Transport)
					continue
				}

				if notiErr != nil {
					logger.WithField("id", snd.ID).WithError(notiErr).Error("Failed to notify")

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
					if err := n.dao.UpdateApprovement(snd); err != nil {
						logger.
							WithError(err).
							WithField("id", snd.ID).
							Error("Failed to update request")
						sleep(time.Second * 30)
						out = true
						continue
					}
				} else {
					logger.WithField("id", snd.ID).Debug("Notified")
				}
			}
		}
	}()

	// sendings
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger := n.logger.WithField("sendings", "")

		for {
			select {
			case <-stopper:
				return
			default:
			}

			// get list
			list, err := n.dao.ListUnnotifiedSendings(itemsPerShot)
			if err != nil {
				logger.WithError(err).Error("Failed to get unnotified requests")
				sleep(time.Second * 30)
				continue
			}

			// nothing
			if len(list) == 0 {
				sleep(time.Second * 30)
				continue
			}

			out := false
			for _, snd := range list {
				if out {
					break
				}

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
					logger.
						WithError(err).
						WithField("id", snd.ID).
						Error("Failed to update request")
					sleep(time.Second * 30)
					out = true
					continue
				}

				notiErrorDesc := "Transaction failed"
				if snd.Status == types.SendingConfirmed {
					notiErrorDesc = ""
				}

				// notify
				var notiErr error
				switch snd.Transport {
				case types.SendingNats:
					if n.natsTransporter != nil {
						notiErr = n.natsTransporter.PublishSentEvent(
							snd.Status == types.SendingConfirmed,
							notiErrorDesc,
							snd.Service, snd.RequestID,
							snd.To, snd.Token, snd.Amount, snd.Digest,
						)
					} else {
						logger.Warn("Nats transport is disabled, skipping notification")
						continue
					}
				case types.SendingHTTP:
					if n.httpTransporter != nil {
						if snd.CallbackURL != "" {
							notiErr = n.httpTransporter.PublishSentEvent(
								snd.Status == types.SendingConfirmed,
								notiErrorDesc,
								snd.Service, snd.RequestID, snd.CallbackURL,
								snd.To, snd.Token, snd.Amount, snd.Digest,
							)
						}
					} else {
						logger.Warn("HTTP transport is disabled, skipping notification")
						continue
					}
				default:
					logger.Errorf("Transport %v is not implemented", snd.Transport)
					continue
				}

				if notiErr != nil {
					logger.WithField("id", snd.ID).WithError(notiErr).Error("Failed to notify")

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
						logger.
							WithError(err).
							WithField("id", snd.ID).
							Error("Failed to update request")
						sleep(time.Second * 30)
						out = true
						continue
					}
				} else {
					logger.WithField("id", snd.ID).Debug("Notified")
				}
			}
		}
	}()

	// wait interruption
	for !token.Stopped() {
		time.Sleep(time.Second)
	}
	close(stopper)

	wg.Wait()
}
