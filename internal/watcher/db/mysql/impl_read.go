package mysql

import (
	"strings"
	"time"

	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
)

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

// ListUnnotifiedIncomings implementation
func (d *Database) ListUnnotifiedIncomings(max uint16) ([]*types.Incoming, error) {
	m := make([]*model.Incoming, 0)

	res := d.
		Model(&model.Incoming{}).
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
