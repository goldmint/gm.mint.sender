package mysql

import (
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
)

// PutWallet impl.
func (d *Database) PutWallet(v *types.PutWallet) error {
	mlist := make([]*model.Wallet, 0)
	for _, pubkey := range v.PublicKeys {
		mlist = append(
			mlist,
			&model.Wallet{
				Base: model.Base{
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				},
				PublicKey: append([]byte(nil), pubkey[:]...),
			},
		)
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

// DeleteWallet impl.
func (d *Database) DeleteWallet(v *types.DeleteWallet) error {
	tx := d.Begin()
	txok := false
	defer func() {
		if !txok {
			tx.Rollback()
		}
	}()
	for _, pubkey := range v.PublicKeys {
		if err := tx.Delete(&model.Wallet{
			PublicKey: append([]byte(nil), pubkey[:]...),
		}).Error; err != nil {
			return err
		}
	}
	txok = true
	return tx.Commit().Error
}

// PutIncoming impl.
func (d *Database) PutIncoming(v *types.PutIncoming) error {
	m := &model.Incoming{
		Base: model.Base{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		To:        v.To[:],
		From:      v.From[:],
		Amount:    v.Amount.String(),
		Token:     uint16(v.Token),
		Digest:    v.Digest[:],
		Block:     v.Block.Bytes(),
		Timestamp: v.Timestamp,
		Sent:      false,
	}
	if err := d.Create(m).Error; err != nil {
		if !d.DuplicateError(err) {
			return err
		}
	}
	return nil
}

// MarkIncomingSent impl.
func (d *Database) MarkIncomingSent(v *types.MarkIncomingSent) error {
	return d.Model(&model.Incoming{
		Digest: v.Digest[:],
	}).Update("sent", v.Sent).Error
}
