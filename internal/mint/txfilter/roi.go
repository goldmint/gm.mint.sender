package txfilter

import (
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/mint/blockparser"
)

// roiCheck decides to fan the transaction out.
// `roiLock` should be locked at the time of the method call
func (f *Filter) roiCheck(tx *blockparser.Transaction) bool {
	in, out := false, false
	_, out = f.roiWallets[tx.From]
	if tx.To != nil {
		_, in = f.roiWallets[*tx.To]
	}
	if in || out {
		return f.txFilter(tx.Type, out)
	}
	return false
}

// addWallet adds a wallet to the ROI.
// `roiLock` should be locked at the time of the method call
func (f *Filter) addWallet(p mint.PublicKey) bool {
	if _, ok := f.roiWallets[p]; !ok {
		f.roiWallets[p] = struct{}{}

		// metrics
		if f.metrics != nil {
			f.metrics.ROIWallets.Add(1)
		}
		return true
	}
	return false
}

// removeWallet removes a wallet from the ROI.
// `roiLock` should be locked at the time of the method call
func (f *Filter) removeWallet(p mint.PublicKey) bool {
	if _, ok := f.roiWallets[p]; ok {
		delete(f.roiWallets, p)

		// metrics
		if f.metrics != nil {
			f.metrics.ROIWallets.Sub(1)
		}
		return true
	}
	return false
}
