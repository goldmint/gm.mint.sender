package txsaver

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/void616/gm.mint.sender/internal/mint/blockparser"
	"github.com/void616/gm.mint.sender/internal/watcher/api/model"
	"github.com/void616/gm.mint.sender/internal/watcher/db"
	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
	mint "github.com/void616/gm.mint"
)

// Saver saves filtered transactions to the DB
type Saver struct {
	logger         *logrus.Entry
	transactions   <-chan *blockparser.Transaction
	walletSubs     <-chan model.WalletSub
	unfilterWallet chan<- mint.PublicKey
	dao            db.DAO
	subs           map[mint.PublicKey]servicesMap
	subsLock       sync.Mutex
}

type servicesMap map[string]types.Service

// New Saver instance
func New(
	transactions <-chan *blockparser.Transaction,
	walletSubs <-chan model.WalletSub,
	unfilterWallet chan<- mint.PublicKey,
	dao db.DAO,
	logger *logrus.Entry,
) (*Saver, error) {
	f := &Saver{
		logger:         logger,
		transactions:   transactions,
		walletSubs:     walletSubs,
		dao:            dao,
		subs:           make(map[mint.PublicKey]servicesMap),
		unfilterWallet: unfilterWallet,
	}
	return f, nil
}

// AddWalletSubs adds subscribers of the specific wallet
func (s *Saver) AddWalletSubs(p mint.PublicKey, services ...types.Service) {
	s.subsLock.Lock()
	defer s.subsLock.Unlock()
	for _, svc := range services {
		if _, ok := s.subs[p]; !ok {
			s.subs[p] = servicesMap{}
		}
		s.subs[p][svc.Name] = svc
	}
}
