package api

import (
	"github.com/void616/gm-mint-sender/internal/watcher/api/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// AddWallet adds wallet to the DB and sends it to the transaction filter
func (api *API) AddWallet(svc string, trans types.ServiceTransport, svcURL string, pub ...sumuslib.PublicKey) bool {

	if err := api.dao.PutService(&types.Service{
		Name:        svc,
		Transport:   trans,
		CallbackURL: svcURL,
	}); err != nil {
		api.logger.WithError(err).Error("Failed to add service")
		return false
	}

	service, err := api.dao.GetService(svc)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get service")
		return false
	}
	if service == nil {
		api.logger.WithError(err).Error("Failed to find service")
		return false
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   *service,
		}
	}

	if err := api.dao.PutWallet(list...); err != nil {
		api.logger.WithError(err).Error("Failed to add wallets")
		return false
	}
	for _, p := range pub {
		api.walletSubs <- model.WalletSub{
			PublicKey: p,
			Service:   *service,
			Add:       true,
		}
		api.watchWallet <- p
	}
	return true
}

// RemoveWallet removes wallet from the DB and from the transaction filter
func (api *API) RemoveWallet(svc string, pub ...sumuslib.PublicKey) bool {

	service, err := api.dao.GetService(svc)
	if err != nil {
		api.logger.WithError(err).Error("Failed to get service")
		return false
	}
	if service == nil {
		api.logger.WithError(err).Error("Failed to find service")
		return false
	}

	list := make([]*types.Wallet, len(pub))
	for i, v := range pub {
		list[i] = &types.Wallet{
			PublicKey: v,
			Service:   *service,
		}
	}

	if err := api.dao.DeleteWallet(list...); err != nil {
		api.logger.WithError(err).Error("Failed to remove wallets")
		return false
	}
	for _, p := range pub {
		api.walletSubs <- model.WalletSub{
			PublicKey: p,
			Service:   *service,
			Add:       false,
		}
	}
	return true
}
