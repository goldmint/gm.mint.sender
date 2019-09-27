package mysql

import (
	"fmt"
	"strconv"
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
	services := make([]*model.Service, 0)
	if err := d.Model(&model.Service{}).Find(&services).Error; err != nil {
		return nil, err
	}
	mlist := make([]struct {
		PublicKey []byte
		ServiceID string
	}, 0)
	table := d.NewScope(&model.Wallet{}).QuotedTableName()
	if err := d.Raw("SELECT `public_key`, GROUP_CONCAT(`service_id` SEPARATOR '|') AS `service_id` FROM " + table + " GROUP BY `public_key`").Scan(&mlist).Error; err != nil {
		return nil, err
	}
	list := make([]*types.WalletServices, len(mlist))
	for i, m := range mlist {
		pub, err := sumuslib.BytesToPublicKey(m.PublicKey)
		if err != nil {
			return nil, err
		}
		svcs := make([]types.Service, 0)
		for _, idstr := range strings.Split(m.ServiceID, "|") {
			id, err := strconv.ParseUint(strings.TrimSpace(idstr), 10, 64)
			if err != nil {
				return nil, err
			}
			found := false
			for _, svc := range services {
				if svc.ID == id {
					msvc, err := svc.MapTo()
					if err != nil {
						return nil, err
					}
					svcs = append(svcs, *msvc)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("failed to find service #%v", id)
			}
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
		if err := tx.Delete(&model.Wallet{}, "`public_key`=? AND `service_id`=?", m.PublicKey, m.Service.ID).Error; err != nil {
			return err
		}
	}
	txok = true
	return tx.Commit().Error
}
