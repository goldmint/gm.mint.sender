package txfilter

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/blockparser"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gotask"
)

// Task loop
func (f *Filter) Task(token *gotask.Token) {
	buf := make([]*blockparser.Transaction, filterBufferSize)

	f.logger.Debugf("%v wallets within ROI", len(f.roiWallets))

	for !token.Stopped() || len(buf) != 0 {
		f.roiLock.Lock()

		// get incoming parsed transactions, filter
		{
			buf = buf[0:0]
			leave := false
			gotTotal := 0
			for !leave && len(buf) < cap(buf) {
				select {
				case tx := <-f.in:
					gotTotal++
					if f.roiCheck(tx) {
						buf = append(buf, tx)
					}
				case <-time.After(time.Second):
					leave = true
				}
			}
			if gotTotal > 0 {
				f.logger.Debugf("Filtered %v from %v transactions", len(buf), gotTotal)
			}
		}

		// flush filtered transactions
		{
			if len(buf) > 0 {
				// metrics
				if f.mtxTxVolumeCounter != nil {
					for _, tx := range buf {
						f.mtxTxVolumeCounter.WithLabelValues("gold").Add(tx.AmountGOLD.Float64())
						f.mtxTxVolumeCounter.WithLabelValues("mnt").Add(tx.AmountMNT.Float64())
					}
				}

				// metrics
				if f.mtxQueueGauge != nil {
					f.mtxQueueGauge.WithLabelValues("txfilter_flush").Set(float64(len(buf)))
				}
				t := time.Now()

				for _, tx := range buf {
					f.out <- tx
				}

				// metrics
				if f.mtxTaskDuration != nil {
					f.mtxTaskDuration.WithLabelValues("txfilter_flush").Observe(time.Since(t).Seconds())
				}
				if f.mtxQueueGauge != nil {
					f.mtxQueueGauge.WithLabelValues("txfilter_flush").Set(0)
				}

				f.logger.Debugf("Flushed %v transactions", len(buf))
			}
		}

		// add wallets to roi
		{
			leave := false
			for !leave {
				select {
				case pubkey := <-f.add:
					f.addWallet(pubkey)
					f.logger.Debugf("Wallet %v added to ROI", sumuslib.Pack58(pubkey[:]))
				default:
					leave = true
				}
			}
		}

		// remove wallets from roi
		{
			leave := false
			for !leave {
				select {
				case pubkey := <-f.remove:
					f.removeWallet(pubkey)
					f.logger.Debugf("Wallet %v removed from ROI", sumuslib.Pack58(pubkey[:]))
				default:
					leave = true
				}
			}
		}

		f.roiLock.Unlock()
	}
}
