package walletservice

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// AddWallet adds wallet to the DB and sends it to the transaction filter
func (s *Service) AddWallet(service string, pub ...sumuslib.PublicKey) bool {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "add_wallet")
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   service,
		}
	}

	if err := s.dao.PutWallet(list...); err != nil {
		s.logger.WithError(err).Error("Failed to add wallets")
		return false
	}
	for _, p := range pub {
		s.walletSubs <- WalletSub{
			PublicKey: p,
			Service:   service,
			Add:       true,
		}
		s.watchWallet <- p
	}
	return true
}

// RemoveWallet removes wallet from the DB and from the transaction filter
func (s *Service) RemoveWallet(service string, pub ...sumuslib.PublicKey) bool {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "remove_wallet")
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   service,
		}
	}

	if err := s.dao.DeleteWallet(list...); err != nil {
		s.logger.WithError(err).Error("Failed to remove wallets")
		return false
	}
	for _, p := range pub {
		s.walletSubs <- WalletSub{
			PublicKey: p,
			Service:   service,
			Add:       false,
		}
	}
	return true
}
