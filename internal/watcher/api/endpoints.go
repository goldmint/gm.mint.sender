package api

import (
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/watcher/api/model"
	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
)

// AddWallet adds wallet to the DB and sends it to the transaction filter
func (api *API) AddWallet(trans types.ServiceTransport, service, callbackURL string, pub ...mint.PublicKey) bool {

	if err := api.dao.PutService(&types.Service{
		Name:        service,
		Transport:   trans,
		CallbackURL: callbackURL,
	}); err != nil {
		api.logger.WithError(err).Error("Failed to add service")
		return false
	}

	s, err := api.dao.GetService(service)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get service")
		return false
	}
	if s == nil {
		api.logger.WithError(err).Error("Failed to find service")
		return false
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   *s,
		}
	}

	if err := api.dao.PutWallet(list...); err != nil {
		api.logger.WithError(err).Error("Failed to add wallets")
		return false
	}
	for _, p := range pub {
		api.walletSubs <- model.WalletSub{
			PublicKey: p,
			Service:   *s,
			Add:       true,
		}
		api.watchWallet <- p
	}
	return true
}

// RemoveWallet removes wallet from the DB and from the transaction filter
func (api *API) RemoveWallet(service string, pub ...mint.PublicKey) bool {

	s, err := api.dao.GetService(service)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get service")
		return false
	}
	if s == nil {
		api.logger.WithError(err).Error("Failed to find service")
		return false
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   *s,
		}
	}

	if err := api.dao.DeleteWallet(list...); err != nil {
		api.logger.WithError(err).Error("Failed to remove wallets")
		return false
	}
	for _, p := range pub {
		api.walletSubs <- model.WalletSub{
			PublicKey: p,
			Service:   *s,
			Add:       false,
		}
	}
	return true
}
