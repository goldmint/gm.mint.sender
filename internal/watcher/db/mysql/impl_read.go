package mysql

import (
	"math/big"

	"github.com/void616/gm-mint-sender/internal/watcher/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/watcher/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// ListWallets impl.
func (d *Database) ListWallets() ([]*types.ListWalletsItem, error) {
	m := make([]*model.Wallet, 0)
	res := d.Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.ListWalletsItem, len(m))
	for i, v := range m {
		var pubkey sumuslib.PublicKey
		copy(pubkey[:], v.PublicKey)
		list[i] = &types.ListWalletsItem{
			PublicKey: pubkey,
		}
	}
	return list, nil
}

// ListUnsentIncomings impl.
func (d *Database) ListUnsentIncomings(max uint16) ([]*types.ListUnsentIncomingsItem, error) {
	m := make([]*model.Incoming, 0)
	res := d.Where("`sent`=0").Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.ListUnsentIncomingsItem, len(m))
	for i, v := range m {
		var vTo sumuslib.PublicKey
		copy(vTo[:], v.To)
		var vFrom sumuslib.PublicKey
		copy(vFrom[:], v.From)
		var vDigest sumuslib.Digest
		copy(vDigest[:], v.Digest)
		list[i] = &types.ListUnsentIncomingsItem{
			To:        vTo,
			From:      vFrom,
			Amount:    amount.NewFloatString(v.Amount),
			Token:     sumuslib.Token(v.Token),
			Digest:    vDigest,
			Block:     big.NewInt(0).SetBytes(v.Block),
			Timestamp: v.Timestamp,
		}
	}
	return list, nil
}
