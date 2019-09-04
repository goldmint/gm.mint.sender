package txfilter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/mint/blockparser"
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
	metrics    *Metrics
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
	logger *logrus.Entry,
) (*Filter, error) {
	f := &Filter{
		logger:     logger,
		in:         in,
		out:        out,
		add:        add,
		remove:     remove,
		roiWallets: make(map[sumuslib.PublicKey]struct{}),
		txFilter:   txFilter,
	}
	return f, nil
}

// AddWallet adds a wallet to the ROI and should be called before service launch
func (f *Filter) AddWallet(pubkey ...sumuslib.PublicKey) {
	f.roiLock.Lock()
	defer f.roiLock.Unlock()
	for _, p := range pubkey {
		f.addWallet(p)
	}
}

// Metrics data
type Metrics struct {
	ROIWallets prometheus.Gauge
	TxVolume   *prometheus.CounterVec
}

// AddMetrics adds metrics counters and should be called before service launch
func (f *Filter) AddMetrics(m *Metrics) {
	f.metrics = m
}
