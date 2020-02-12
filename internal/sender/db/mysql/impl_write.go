package mysql

import (
	"math/big"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/db/mysql/model"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
)

// PutWallet implementation
func (d *Database) PutWallet(v *types.Wallet) error {
	m := &model.Wallet{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	if err := d.Create(m).Error; err != nil {
		if !d.DuplicateError(err) {
			return err
		}
	}
	return nil
}

// PutSending implementation
func (d *Database) PutSending(v *types.Sending) error {
	m := &model.Sending{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	if err := d.Create(m).Error; err != nil {
		return err
	}
	v.ID = m.ID
	return nil
}

// UpdateSending implementation
func (d *Database) UpdateSending(v *types.Sending) error {
	var m = &model.Sending{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	return d.Save(m).Error
}

// SetSendingConfirmed implementation
func (d *Database) SetSendingConfirmed(dig mint.Digest, from mint.PublicKey, block *big.Int) error {
	return d.Model(&model.Sending{}).
		Where("`digest`=? AND `sender`=?", dig.Bytes(), from.Bytes()).
		Update(
			map[string]interface{}{
				"status": uint8(types.SendingConfirmed),
				"block":  block.Bytes(),
			},
		).
		Limit(1).
		Error
}

// PutApprovement implementation
func (d *Database) PutApprovement(v *types.Approvement) error {
	m := &model.Approvement{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	if err := d.Create(m).Error; err != nil {
		return err
	}
	v.ID = m.ID
	return nil
}

// UpdateApprovement implementation
func (d *Database) UpdateApprovement(v *types.Approvement) error {
	var m = &model.Approvement{}
	if err := m.MapFrom(v); err != nil {
		return err
	}
	return d.Save(m).Error
}

// SetApprovementConfirmed implementation
func (d *Database) SetApprovementConfirmed(dig mint.Digest, from mint.PublicKey, block *big.Int) error {
	return d.Model(&model.Approvement{}).
		Where("`digest`=? AND `sender`=?", dig.Bytes(), from.Bytes()).
		Update(
			map[string]interface{}{
				"status": uint8(types.SendingConfirmed),
				"block":  block.Bytes(),
			},
		).
		Limit(1).
		Error
}
