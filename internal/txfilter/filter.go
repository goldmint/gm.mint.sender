package txfilter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	sumuslib "github.com/void616/gm-sumuslib"
)

const filterBufferSize = 64

// Filter filters parsed txs by ROI-wallets and tx type
type Filter struct {
	logger     *logrus.Entry
	in         <-chan *blockparser.Transaction
	out        chan<- *blockparser.Transaction
	add        <-chan sumuslib.PublicKey
	remove     <-chan sumuslib.PublicKey
	roiLock    sync.Mutex
	roiWallets map[sumuslib.PublicKey]struct{}
	txFilter   TxFilter

	mtxROIWalletsGauge prometheus.Gauge
	mtxTxVolumeCounter *prometheus.CounterVec
	mtxTaskDuration    *prometheus.SummaryVec
	mtxQueueGauge      *prometheus.GaugeVec
}

// TxFilter filters transaction
type TxFilter func(typ sumuslib.Transaction, outgoing bool) bool

// New Filter instance.
// `in` channel receives transactions. `out` channel emits filtered transactions.
// `add` channel adds a wallet to the ROI. `remove` channel removes a wallet from the ROI.
// `txFilter` filters transactions by type
func New(
	in <-chan *blockparser.Transaction,
	out chan<- *blockparser.Transaction,
	add <-chan sumuslib.PublicKey,
	remove <-chan sumuslib.PublicKey,
	txFilter TxFilter,
	mtxROIWalletsGauge prometheus.Gauge,
	mtxTxVolumeCounter *prometheus.CounterVec,
	mtxTaskDuration *prometheus.SummaryVec,
	mtxQueueGauge *prometheus.GaugeVec,
	logger *logrus.Entry,
) (*Filter, error) {
	f := &Filter{
		logger:             logger,
		in:                 in,
		out:                out,
		add:                add,
		remove:             remove,
		roiWallets:         make(map[sumuslib.PublicKey]struct{}),
		txFilter:           txFilter,
		mtxROIWalletsGauge: mtxROIWalletsGauge,
		mtxTxVolumeCounter: mtxTxVolumeCounter,
		mtxTaskDuration:    mtxTaskDuration,
		mtxQueueGauge:      mtxQueueGauge,
	}
	return f, nil
}

// AddWallet adds a wallet to the ROI.
// Should be used to add all known wallets before filter is started
func (f *Filter) AddWallet(pubkey ...sumuslib.PublicKey) {
	f.roiLock.Lock()
	defer f.roiLock.Unlock()
	for _, p := range pubkey {
		f.addWallet(p)
	}
}
