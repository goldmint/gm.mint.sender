package txsaver

import (
	"math/big"
	"time"

	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gotask"
)

// Task loop
func (s *Saver) Task(token *gotask.Token) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()

	s.logger.Debugf("%v wallets within ROI", len(s.subs))

	empty := false
	zero := big.NewInt(0)

	for !token.Stopped() || !empty {
		empty = false
		savedItems := 0
		for !empty {
			select {

			// save next filtered transaction
			case tx := <-s.transactions:
				// asset transaction
				if tx.Type != mint.TransactionTransferAssets {
					break
				}

				// destination is known
				if tx.To == nil {
					break
				}

				// some coins are transferred
				var tkn mint.Token
				var amo *amount.Amount
				switch {
				case tx.AmountMNT.Value.Cmp(zero) > 0:
					tkn = mint.TokenMNT
					amo = amount.FromAmount(tx.AmountMNT)
				case tx.AmountGOLD.Value.Cmp(zero) > 0:
					tkn = mint.TokenGOLD
					amo = amount.FromAmount(tx.AmountGOLD)
				}
				if amo == nil {
					break
				}

				// destination has subscribed services
				if svcs, ok := s.subs[*tx.To]; !ok || len(svcs) == 0 {
					break
				}

				// incoming per service
				models := make([]*types.Incoming, len(s.subs[*tx.To]))
				{
					i := 0
					for _, svc := range s.subs[*tx.To] {
						models[i] = &types.Incoming{
							Service:   svc,
							To:        *tx.To,
							From:      tx.From,
							Amount:    amo,
							Token:     tkn,
							Digest:    tx.Digest,
							Block:     tx.Block,
							Timestamp: tx.Timestamp,
						}
						i++
					}
				}

				// save to death
				saved := false
				for !token.Stopped() && !saved {
					if err := s.dao.PutIncoming(models...); err != nil {
						s.logger.WithError(err).WithField("digest", tx.Digest.String()).Errorf("Failed to save transaction")
						token.Sleep(time.Second * 10)
					} else {
						saved = true
						savedItems++
					}
				}

			// add/remove wallet:service pair
			case pair := <-s.walletSubs:
				// add
				if pair.Add {
					if _, ok := s.subs[pair.PublicKey]; !ok {
						s.subs[pair.PublicKey] = servicesMap{}
					}
					if _, ok := s.subs[pair.PublicKey][pair.Service.Name]; !ok {
						s.subs[pair.PublicKey][pair.Service.Name] = pair.Service
						s.logger.Debugf("Pair %v:%v added to ROI", pair.PublicKey.StringMask(), pair.Service.Name)
					}
					break
				}
				// remove
				if _, ok := s.subs[pair.PublicKey]; ok {
					if _, ok1 := s.subs[pair.PublicKey][pair.Service.Name]; ok1 {
						delete(s.subs[pair.PublicKey], pair.Service.Name)
						// no more services => don't filter tx-s at all for this address
						if len(s.subs[pair.PublicKey]) == 0 {
							s.unfilterWallet <- pair.PublicKey
						}
						s.logger.Debugf("Pair %v:%v removed from ROI", pair.PublicKey.StringMask(), pair.Service.Name)
					}
				}

			// nothing to do
			case <-time.After(time.Millisecond * 250):
				empty = true
			}
		}
		if savedItems > 0 {
			s.logger.Debugf("Saved %v transactions", savedItems)
		}
	}
}
