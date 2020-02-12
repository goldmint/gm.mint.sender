package mysql

import (
	"math/big"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/db/mysql/model"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
)

// ListWallets implementation
func (d *Database) ListWallets() ([]*types.Wallet, error) {
	m := make([]*model.Wallet, 0)
	res := d.Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.Wallet, len(m))
	for i, v := range m {
		w, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = w
	}
	return list, nil
}

// ListEnqueuedSendings implementation
func (d *Database) ListEnqueuedSendings(max uint16) ([]*types.Sending, error) {
	m := make([]*model.Sending, 0)
	res := d.Where("`status`=?", uint8(types.SendingEnqueued)).Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.Sending, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// ListStaleSendings implementation
func (d *Database) ListStaleSendings(elderThanBlockID *big.Int, max uint16) ([]*types.Sending, error) {
	m := make([]*model.Sending, 0)
	res := d.Where(
		"`status`=? AND `sent_at_block` IS NOT NULL AND `sent_at_block`<?",
		uint8(types.SendingPosted),
		elderThanBlockID.Bytes(),
	).
		Limit(max).
		Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.Sending, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// ListUnnotifiedSendings implementation
func (d *Database) ListUnnotifiedSendings(max uint16) ([]*types.Sending, error) {
	m := make([]*model.Sending, 0)

	res := d.
		Model(&model.Sending{}).
		Where(
			"(`status`=? OR `status`=?) AND `notified`=0 AND (`notify_at` IS NULL OR `notify_at`<=?)",
			uint8(types.SendingConfirmed),
			uint8(types.SendingFailed),
			time.Now().UTC(),
		).
		Limit(max).
		Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}

	list := make([]*types.Sending, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// ListEnqueuedApprovements implementation
func (d *Database) ListEnqueuedApprovements(max uint16) ([]*types.Approvement, error) {
	m := make([]*model.Approvement, 0)
	res := d.Where("`status`=?", uint8(types.SendingEnqueued)).Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.Approvement, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// ListStaleApprovements implementation
func (d *Database) ListStaleApprovements(elderThanBlockID *big.Int, max uint16) ([]*types.Approvement, error) {
	m := make([]*model.Approvement, 0)
	res := d.Where(
		"`status`=? AND `sent_at_block` IS NOT NULL AND `sent_at_block`<?",
		uint8(types.SendingPosted),
		elderThanBlockID.Bytes(),
	).
		Limit(max).
		Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.Approvement, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// ListUnnotifiedApprovements implementation
func (d *Database) ListUnnotifiedApprovements(max uint16) ([]*types.Approvement, error) {
	m := make([]*model.Approvement, 0)

	res := d.
		Model(&model.Approvement{}).
		Where(
			"(`status`=? OR `status`=?) AND `notified`=0 AND (`notify_at` IS NULL OR `notify_at`<=?)",
			uint8(types.SendingConfirmed),
			uint8(types.SendingFailed),
			time.Now().UTC(),
		).
		Limit(max).
		Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}

	list := make([]*types.Approvement, len(m))
	for i, v := range m {
		s, err := v.MapTo()
		if err != nil {
			return nil, err
		}
		list[i] = s
	}
	return list, nil
}

// EarliestBlock implementation
func (d *Database) EarliestBlock() (*big.Int, bool, error) {

	var min = new(big.Int)
	var valid = false
	var check = func(x *big.Int) {
		if !valid {
			min.Set(x)
			valid = true
			return
		}
		if x.Cmp(min) < 0 {
			min.Set(x)
		}
	}

	// check sendings
	{
		m := struct {
			Earliest []byte
		}{}
		res := d.
			Table(d.tablePrefix+"sendings").
			Select("MIN(`sent_at_block`) as `earliest`").
			Where("`status`=?", uint8(types.SendingPosted)).
			First(&m)
		if res.Error != nil {
			return nil, false, res.Error
		}
		if len(m.Earliest) != 0 {
			check(new(big.Int).SetBytes(m.Earliest))
		}
	}

	// check approvements
	{
		m := struct {
			Earliest []byte
		}{}
		res := d.
			Table(d.tablePrefix+"approvements").
			Select("MIN(`sent_at_block`) as `earliest`").
			Where("`status`=?", uint8(types.SendingPosted)).
			First(&m)
		if res.Error != nil {
			return nil, false, res.Error
		}
		if len(m.Earliest) != 0 {
			check(new(big.Int).SetBytes(m.Earliest))
		}
	}

	return min, valid, nil
}

// LatestSenderNonce implementation
func (d *Database) LatestSenderNonce(sender mint.PublicKey) (uint64, error) {

	var max uint64
	var check = func(x uint64) {
		if x > max {
			x = max
		}
	}

	// sendings
	{
		m := struct {
			Latest uint64
		}{}
		res := d.
			Table(d.tablePrefix+"sendings").
			Select("MAX(`sender_nonce`) as `latest`").
			Where("`sender`=?", sender[:]).
			First(&m)
		if res.Error != nil {
			return 0, res.Error
		}
		if m.Latest > 0 {
			check(m.Latest)
		}
	}

	// approvements
	{
		m := struct {
			Latest uint64
		}{}
		res := d.
			Table(d.tablePrefix+"approvements").
			Select("MAX(`sender_nonce`) as `latest`").
			Where("`sender`=?", sender[:]).
			First(&m)
		if res.Error != nil {
			return 0, res.Error
		}
		if m.Latest > 0 {
			check(m.Latest)
		}
	}

	return max, nil
}
