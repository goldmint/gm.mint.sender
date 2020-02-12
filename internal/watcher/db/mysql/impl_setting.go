package mysql

import (
	"github.com/void616/gm.mint.sender/internal/watcher/db/mysql/model"
)

// PutSetting implementation
func (d *Database) PutSetting(k, v string) error {
	m := &model.Setting{
		Key:   k,
		Value: model.LimitStringField(v, 1024), // see model
	}
	if err := d.Create(m).Error; err != nil {
		if !d.DuplicateError(err) {
			return err
		}
		return d.Model(&model.Setting{}).Update(m).Error
	}
	return nil
}

// GetSetting implementation
func (d *Database) GetSetting(k, def string) (string, error) {
	m := &model.Setting{}
	res := d.Model(&model.Setting{}).Where("`key`=?", k).First(m)
	if res.RecordNotFound() {
		return def, nil
	}
	if res.Error != nil {
		return "", res.Error
	}
	return m.Value, nil
}
