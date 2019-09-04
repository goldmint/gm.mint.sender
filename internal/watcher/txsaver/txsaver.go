package txsaver

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm-mint-sender/internal/mint/blockparser"
	"github.com/void616/gm-mint-sender/internal/watcher/api/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db"
	sumuslib "github.com/void616/gm-sumuslib"
)

// Saver saves filtered transactions to the DB
type Saver struct {
	logger         *logrus.Entry
	transactions   <-chan *blockparser.Transaction
	walletSubs     <-chan model.WalletSub
	unfilterWallet chan<- sumuslib.PublicKey
	dao            db.DAO
	subs           map[sumuslib.PublicKey]submap
	subsLock       sync.Mutex
}

type submap map[string]struct{}

// New Saver instance
func New(
	transactions <-chan *blockparser.Transaction,
	walletSubs <-chan model.WalletSub,
	unfilterWallet chan<- sumuslib.PublicKey,
	dao db.DAO,
	logger *logrus.Entry,
) (*Saver, error) {
	f := &Saver{
		logger:         logger,
		transactions:   transactions,
		walletSubs:     walletSubs,
		dao:            dao,
		subs:           make(map[sumuslib.PublicKey]submap),
		unfilterWallet: unfilterWallet,
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
