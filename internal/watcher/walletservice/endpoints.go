package walletservice

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// AddWallet adds wallet to the DB and sends it to the transaction filter
func (s *Service) AddWallet(pub ...sumuslib.PublicKey) bool {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "add_wallet")
	}

	if err := s.dao.PutWallet(&types.PutWallet{
		PublicKeys: pub,
	}); err != nil {
		s.logger.WithError(err).Error("Failed to add wallets")
		return false
	}
	for _, p := range pub {
		s.addWallet <- p
	}
	return true
}

// RemoveWallet removes wallet from the DB and from the transaction filter
func (s *Service) RemoveWallet(pub ...sumuslib.PublicKey) bool {
	// metrics
	if s.mtxMethodDuration != nil {
		defer func(t time.Time, method string) {
			s.mtxMethodDuration.WithLabelValues(method).Observe(time.Since(t).Seconds())
		}(time.Now(), "remove_wallet")
	}

	if err := s.dao.DeleteWallet(&types.DeleteWallet{
		PublicKeys: pub,
	}); err != nil {
		s.logger.WithError(err).Error("Failed to remove wallets")
		return false
	}
	for _, p := range pub {
		s.removeWallet <- p
	}
	return true
}
