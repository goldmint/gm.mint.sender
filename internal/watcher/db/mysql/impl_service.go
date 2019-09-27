package mysql

import (
	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
)

// PutService implementation
func (d *Database) PutService(v *types.Service) error {
	m := &model.Service{}
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

// GetService implementation
func (d *Database) GetService(name string) (*types.Service, error) {
	m := &model.Service{}
	res := d.Model(&model.Service{}).Where("`name`=?", name).First(m)
	if res.RecordNotFound() {
		return nil, nil
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return m.MapTo()
}
