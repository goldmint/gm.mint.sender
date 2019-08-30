package txsaver

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/blockparser"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	"github.com/void616/gm-mint-sender/internal/watcher/walletservice"
	sumuslib "github.com/void616/gm-sumuslib"
)

// Saver saves filtered transactions to the DB
type Saver struct {
	logger          *logrus.Entry
	transactions    <-chan *blockparser.Transaction
	walletSubs      <-chan walletservice.WalletSub
	unfilterWallet  chan<- sumuslib.PublicKey
	dao             db.DAO
	subs            map[sumuslib.PublicKey]submap
	subsLock        sync.Mutex
	mtxTaskDuration *prometheus.SummaryVec
}

type submap map[string]struct{}

// New Saver instance
func New(
	transactions <-chan *blockparser.Transaction,
	walletSubs <-chan walletservice.WalletSub,
	unfilterWallet chan<- sumuslib.PublicKey,
	dao db.DAO,
	mtxTaskDuration *prometheus.SummaryVec,
	logger *logrus.Entry,
) (*Saver, error) {
	f := &Saver{
		logger:          logger,
		transactions:    transactions,
		walletSubs:      walletSubs,
		dao:             dao,
		subs:            make(map[sumuslib.PublicKey]submap),
		unfilterWallet:  unfilterWallet,
		mtxTaskDuration: mtxTaskDuration,
	}
	return f, nil
}

// AddWalletSubs adds subscribers of the specific wallet
func (s *Saver) AddWalletSubs(p sumuslib.PublicKey, service ...string) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()
	for _, svc := range service {
		if svc != "" {
			if _, ok := s.subs[p]; !ok {
				s.subs[p] = submap{}
			}
			s.subs[p][svc] = struct{}{}
		}
	}
}
