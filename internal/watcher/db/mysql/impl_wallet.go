package mysql

import (
	"strings"

	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

// PutWallet implementation
func (d *Database) PutWallet(v ...*types.Wallet) error {
	mlist := make([]*model.Wallet, 0)
	for _, w := range v {
		m := &model.Wallet{}
		if err := m.MapFrom(w); err != nil {
			return err
		}
		mlist = append(mlist, m)
	}
	tx := d.Begin()
	txok := false
	defer func() {
		if !txok {
			tx.Rollback()
		}
	}()
	for _, m := range mlist {
		if err := tx.Create(m).Error; err != nil {
			if !d.DuplicateError(err) {
				return err
			}
		}
	}
	txok = true
	return tx.Commit().Error
}

// ListWallets implementation
func (d *Database) ListWallets() ([]*types.WalletServices, error) {
	mlist := make([]struct {
		PublicKey []byte
		Service   string
	}, 0)
	table := d.NewScope(&model.Wallet{}).QuotedTableName()
	if err := d.Raw("SELECT `public_key`, GROUP_CONCAT(`service`) AS `service` FROM " + table + " GROUP BY `public_key`").Scan(&mlist).Error; err != nil {
		return nil, err
	}
	list := make([]*types.WalletServices, len(mlist))
	for i, m := range mlist {
		pub, err := sumuslib.BytesToPublicKey(m.PublicKey)
		if err != nil {
			return nil, err
		}
		svcs := strings.Split(m.Service, ",")
		for j := range svcs {
			svcs[j] = strings.TrimSpace(svcs[j])
		}
		list[i] = &types.WalletServices{
			PublicKey: pub,
			Services:  svcs,
		}
	}
	return list, nil
}

// DeleteWallet implementation
func (d *Database) DeleteWallet(v ...*types.Wallet) error {
	mlist := make([]*model.Wallet, 0)
	for _, w := range v {
		m := &model.Wallet{}
		if err := m.MapFrom(w); err != nil {
			return err
		}
		mlist = append(mlist, m)
	}
	tx := d.Begin()
	txok := false
	defer func() {
		if !txok {
			tx.Rollback()
		}
	}()
	for _, m := range mlist {
		if err := tx.Delete(&model.Wallet{}, "`public_key`=? AND `service`=?", m.PublicKey, m.Service).Error; err != nil {
			return err
		}
	}
	txok = true
	return tx.Commit().Error
}
