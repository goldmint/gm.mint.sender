package blockobserver

import (
	"context"
	"math/big"
	"time"

	"github.com/void616/gotask"
)

// Task loop
func (o *Observer) Task(token *gotask.Token) {
	from := new(big.Int).Set(o.from)

	for !token.Stopped() {

		// get connection from the pool
		conn, connClose, err := o.rpcpool.ConnOnly()
		if err != nil {
			o.logger.WithError(err).Error("Failed to get a free connection")
			token.Sleep(time.Second * 10)
			continue
		}

		// connection heartbeat
		if err := conn.Heartbeat(time.Second * 5); err != nil {
			o.logger.WithError(err).Error("Failed to check node connection")
			connClose()
			token.Sleep(time.Second * 3)
			continue
		}

		if token.Stopped() {
			connClose()
			break
		}

		// receiving context
		rctx, ctxCancel := conn.Receive(context.Background())
		go func() {
			defer ctxCancel()

			for {
				evt, err := conn.ReceiveEvent(rctx, "blocks_synchronized")
				if err != nil {
					if err != context.Canceled {
						o.logger.WithError(err).Errorf("Failed to get node event")
					}
					return
				}

				// parse
				latest, err := o.parseEvent(evt)
				if err != nil {
					o.logger.WithError(err).Error("Failed to parse node event")
					return
				}

				// check block id
				if latest.Cmp(big.NewInt(-1)) > 0 {
					o.logger.Warningf("Latest block from listener: %v", latest.String())

					for !token.Stopped() && from.Cmp(latest) < 0 {
						cur := new(big.Int).Set(from)
						cur.Add(cur, big.NewInt(1))

						if err := o.parser.Parse(cur); err != nil {
							o.logger.WithError(err).WithField("block", cur.String()).Error("Failed to parse block")
							token.Sleep(time.Second * 10)
							continue
						}
						o.logger.WithField("block", cur.String()).Debugf("Block completed")

						from.Set(cur)
					}
				}
			}
		}()

		for {
			select {
			case <-rctx.Done():
			case <-time.After(time.Millisecond * 250):
				if !token.Stopped() {
					continue
				}
			}
			break
		}

		// release context
		ctxCancel()
		// release connection
		connClose()
	}
}
