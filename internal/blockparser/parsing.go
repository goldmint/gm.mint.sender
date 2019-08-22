package blockparser

import (
	"fmt"
	"io"
	"math/big"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
	"github.com/void616/gm-sumuslib/block"
	"github.com/void616/gm-sumuslib/serializer"
	"github.com/void616/gm-sumuslib/transaction"
)

// Invoked when transaction-type-specific data should fill transaction model
type ptxCbk func(t *Transaction)

// parseBlockData parses block header and transactions from bytes
func (p *Parser) parseBlockData(r io.Reader) error {
	var blockModel *Block
	return block.Parse(
		r,
		// header parsed
		func(h *block.Header) error {
			signers := make([]sumuslib.PublicKey, 0)
			for _, s := range h.Signers {
				signers = append(signers, s.PublicKey)
			}
			blockModel = &Block{
				Block:             big.NewInt(0).Set(h.BlockNumber),
				PrevDigest:        h.PrevBlockDigest,
				MerkleRoot:        h.MerkleRoot,
				TransactionsCount: h.TransactionsCount,
				Signers:           signers,
				TotalMNT:          amount.New(),
				TotalGOLD:         amount.New(),
				FeeMNT:            amount.New(),
				FeeGOLD:           amount.New(),
				TotalUserData:     0,
				Timestamp:         sumuslib.DateFromStamp(h.Timestamp),
			}
			return nil
		},
		// next transaction is ready to be parsed
		func(t sumuslib.Transaction, d *serializer.Deserializer, h *block.Header) error {
			switch t {

			case sumuslib.TransactionRegisterNode:
				tx := transaction.RegisterNode{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						m.Data = []byte(tx.NodeAddress)
					},
				)

			case sumuslib.TransactionUnregisterNode:
				tx := transaction.UnregisterNode{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						// nothing
					},
				)

			case sumuslib.TransactionTransferAssets:
				tx := transaction.TransferAsset{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						switch tx.Token {
						case sumuslib.TokenMNT:
							m.AmountMNT = amount.NewAmount(tx.Amount)
							// stat
							blockModel.TotalMNT.Value.Add(blockModel.TotalMNT.Value, tx.Amount.Value)
						case sumuslib.TokenGOLD:
							m.AmountGOLD = amount.NewAmount(tx.Amount)
							// stat
							blockModel.TotalGOLD.Value.Add(blockModel.TotalGOLD.Value, tx.Amount.Value)
						}
					},
				)

			case sumuslib.TransactionRegisterSystemWallet:
				tx := transaction.RegisterSysWallet{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						m.Data = []byte(tx.Tag.String())
					},
				)

			case sumuslib.TransactionUnregisterSystemWallet:
				tx := transaction.UnregisterSysWallet{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						m.Data = []byte(tx.Tag.String())
					},
				)

			case sumuslib.TransactionUserData:
				tx := transaction.UserData{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						m.Data = tx.Data
						// stat
						blockModel.TotalUserData += uint64(len(m.Data))
					},
				)

			case sumuslib.TransactionDistributionFee:
				tx := transaction.DistributionFee{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.OwnerAddress
						m.To = &to
						m.AmountMNT = amount.NewAmount(tx.AmountMNT)
						m.AmountGOLD = amount.NewAmount(tx.AmountGOLD)
						// stat
						blockModel.FeeMNT.Value.Add(blockModel.FeeMNT.Value, tx.AmountMNT.Value)
						blockModel.FeeGOLD.Value.Add(blockModel.FeeGOLD.Value, tx.AmountGOLD.Value)
					},
				)
			}

			return fmt.Errorf("Transaction `%v` not implemented in parser", t)
		},
	)
}

// ---

func mkTX(itx transaction.ITransaction, typ sumuslib.Transaction, d *serializer.Deserializer, h *block.Header, pubTX chan<- *Transaction, fillCbk ptxCbk) error {
	ptx, err := itx.Parse(d.Source())
	if err != nil {
		return err
	}
	m := Transaction{
		Digest:     ptx.Digest,
		Block:      big.NewInt(0).Set(h.BlockNumber),
		Type:       typ,
		Nonce:      ptx.Nonce,
		From:       ptx.From,
		To:         nil,
		AmountMNT:  amount.New(),
		AmountGOLD: amount.New(),
		Timestamp:  sumuslib.DateFromStamp(h.Timestamp),
		Data:       nil,
	}
	fillCbk(&m)
	pubTX <- &m
	return nil
}
