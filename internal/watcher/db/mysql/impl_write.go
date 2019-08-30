package mysql

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
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

// PutIncoming implementation
func (d *Database) PutIncoming(v ...*types.Incoming) error {
	mlist := make([]*model.Incoming, 0)
	for _, w := range v {
		m := &model.Incoming{}
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

// UpdateIncoming implementation
func (d *Database) UpdateIncoming(v *types.Incoming) error {
	var m = &model.Incoming{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	return d.Save(m).Error
}
