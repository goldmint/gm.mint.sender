package mysql

import (
	"time"

	"github.com/void616/gm.mint.sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm.mint.sender/internal/watcher/db/types"
)

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

// ListUnnotifiedIncomings implementation
func (d *Database) ListUnnotifiedIncomings(max uint16) ([]*types.Incoming, error) {
	m := make([]*model.Incoming, 0)

	res := d.
		Model(&model.Incoming{}).
		Preload("Service").
		Where(
			"`notified`=0 AND (`notify_at` IS NULL OR `notify_at`<=?)",
			time.Now().UTC(),
		).
		Limit(max).
		Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}

	list := make([]*types.Incoming, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}
