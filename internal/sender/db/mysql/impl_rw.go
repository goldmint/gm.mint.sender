package mysql

import (
	"errors"
	"math/big"

	"github.com/void616/gm-mint-sender/internal/sender/db/mysql/model"
	"github.com/void616/gm-mint-sender/internal/sender/db/types"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

// SaveSenderWallet saves sending wallet to track it's outgoing transactions later
func (d *Database) SaveSenderWallet(pubkey sumuslib.PublicKey) error {
	m := model.Wallet{
		PublicKey: pubkey[:],
	}
	if err := d.Create(&m).Error; err != nil {
		if !d.DuplicateError(err) {
			return err
		}
	}
	return nil
}

// ListSenderWallets get s list of all known senders
func (d *Database) ListSenderWallets() ([]sumuslib.PublicKey, error) {
	m := make([]*model.Wallet, 0)
	res := d.Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]sumuslib.PublicKey, len(m))
	for i, v := range m {
		var pubkey sumuslib.PublicKey
		copy(pubkey[:], v.PublicKey)
		list[i] = pubkey
	}
	return list, nil
}

// EarliestBlock finds a minimal block ID at which a transaction has been sent
func (d *Database) EarliestBlock() (*types.EarliestBlock, error) {
	m := struct {
		Earliest uint64
	}{}
	res := d.
		Table(d.tablePrefix+"sendings").
		Select("MIN(`sent_at_block`) as `earliest`").
		Where("`status`=?", model.SendingPosted).
		First(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	return &types.EarliestBlock{
		Block: new(big.Int).SetUint64(m.Earliest),
		Empty: m.Earliest == 0,
	}, nil
}

// LatestSenderNonce gets max used nonce for specified sender or zero
func (d *Database) LatestSenderNonce(sender sumuslib.PublicKey) (uint64, error) {
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
	return m.Latest, nil
}

// EnqueueSending adds sending request
func (d *Database) EnqueueSending(request string, to sumuslib.PublicKey, amount *amount.Amount, token sumuslib.Token) error {
	m := model.Sending{
		Status:    model.SendingEnqueued,
		To:        to[:],
		Token:     uint16(token),
		Amount:    amount.String(),
		RequestID: request,
	}
	if err := d.Create(&m).Error; err != nil {
		return err
	}
	return nil
}

// ListEnqueuedSendings gets a list of enqueued sending requests
func (d *Database) ListEnqueuedSendings(max uint16) ([]*types.ListEnqueuedSendingsItem, error) {
	m := make([]*model.Sending, 0)
	res := d.Where("`status`=?", model.SendingEnqueued).Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.ListEnqueuedSendingsItem, len(m))
	for i, v := range m {
		var vTo sumuslib.PublicKey
		copy(vTo[:], v.To)
		list[i] = &types.ListEnqueuedSendingsItem{
			ID:     v.ID,
			To:     vTo,
			Amount: amount.NewFloatString(v.Amount),
			Token:  sumuslib.Token(v.Token),
		}
	}
	return list, nil
}

// SetSendingPosted marks request as posted to the blockchain
func (d *Database) SetSendingPosted(id uint64, sender sumuslib.PublicKey, nonce uint64, digest sumuslib.Digest, block *big.Int) error {
	if !block.IsUint64() {
		return errors.New("can't fit block id into uint64")
	}
	return d.Model(&model.Sending{}).
		Where("`id`=?", id).
		Updates(map[string]interface{}{
			"status":        model.SendingPosted,
			"sender":        sender[:],
			"sender_nonce":  nonce,
			"digest":        digest[:],
			"sent_at_block": block.Uint64(),
		}).
		Limit(1).
		Error
}

// SetSendingFailed marks request as failed
func (d *Database) SetSendingFailed(id uint64) error {
	return d.Model(&model.Sending{}).
		Where("`id`=?", id).
		Updates(map[string]interface{}{
			"status":        model.SendingFailed,
			"sender":        nil,
			"sender_nonce":  nil,
			"digest":        nil,
			"sent_at_block": nil,
		}).
		Limit(1).
		Error
}

// ListStaleSendings gets a list of stale posted requests
func (d *Database) ListStaleSendings(elderThanBlockID *big.Int, max uint16) ([]*types.ListStaleSendingsItem, error) {
	if !elderThanBlockID.IsUint64() {
		return nil, errors.New("can't fit block id into uint64")
	}
	m := make([]*model.Sending, 0)
	res := d.Where("`status`=? AND `sent_at_block` IS NOT NULL AND `sent_at_block`<?", model.SendingPosted, elderThanBlockID.Uint64()).Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.ListStaleSendingsItem, len(m))
	for i, v := range m {
		var vTo sumuslib.PublicKey
		copy(vTo[:], v.To)
		var vFrom sumuslib.PublicKey
		copy(vFrom[:], v.Sender)
		list[i] = &types.ListStaleSendingsItem{
			ID:     v.ID,
			To:     vTo,
			Amount: amount.NewFloatString(v.Amount),
			Token:  sumuslib.Token(v.Token),
			From:   vFrom,
			Nonce:  v.SenderNonce,
		}
	}
	return list, nil
}

// SetSendingConfirmed marks request as confirmed, e.g. fixed on blockchain
func (d *Database) SetSendingConfirmed(sender sumuslib.PublicKey, digest sumuslib.Digest, block *big.Int) error {
	if !block.IsUint64() {
		return errors.New("can't fit block id into uint64")
	}
	return d.Model(&model.Sending{}).
		Where("`sender`=? AND `digest`=? AND `status`=?", sender[:], digest[:], model.SendingPosted).
		Updates(map[string]interface{}{
			"status": model.SendingConfirmed,
			"block":  block.Uint64(),
		}).
		Limit(1).
		Error
}

// ListUnnotifiedSendings gets a list of requests without notification of requestor
func (d *Database) ListUnnotifiedSendings(max uint16) ([]*types.ListUnnotifiedSendingsItem, error) {
	m := make([]*model.Sending, 0)
	res := d.Where("(`status`=? OR `status`=?) AND `notified`=0", model.SendingConfirmed, model.SendingFailed).Limit(max).Find(&m)
	if res.Error != nil {
		return nil, res.Error
	}
	list := make([]*types.ListUnnotifiedSendingsItem, len(m))
	for i, v := range m {
		var vDigest sumuslib.Digest
		copy(vDigest[:], v.Digest)
		list[i] = &types.ListUnnotifiedSendingsItem{
			ID:        v.ID,
			Digest:    vDigest,
			RequestID: v.RequestID,
			Sent:      v.Status != model.SendingFailed,
		}
	}
	return list, nil
}

// SetSendingNotified marks a sending as finally completed (requestor is notified)
func (d *Database) SetSendingNotified(id uint64, notified bool) error {
	var val uint8
	if notified {
		val = 1
	}
	return d.Model(&model.Sending{}).
		Where("`id`=?", id).
		Updates(map[string]interface{}{
			"notified": val,
		}).
		Limit(1).
		Error
}
