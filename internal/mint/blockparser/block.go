package blockparser

import (
	"fmt"
	"io"
	"math/big"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gm.mint/block"
	"github.com/void616/gm.mint/serializer"
	"github.com/void616/gm.mint/transaction"
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
			signers := make([]mint.PublicKey, 0)
			for _, s := range h.Signers {
				signers = append(signers, s.PublicKey)
			}
			blockModel = &Block{
				Block:             big.NewInt(0).Set(h.BlockID),
				PrevDigest:        h.PrevBlockDigest,
				MerkleRoot:        h.MerkleRoot,
				TransactionsCount: h.TransactionsCount,
				Signers:           signers,
				TotalMNT:          amount.New(),
				TotalGOLD:         amount.New(),
				FeeMNT:            amount.New(),
				FeeGOLD:           amount.New(),
				TotalUserData:     0,
				Timestamp:         mint.StampToTime(h.Timestamp),
			}
			return nil
		},
		// next transaction is ready to be parsed
		func(t transaction.Code, d *serializer.Deserializer, h *block.Header) error {
			switch t {

			case transaction.RegisterNodeTx:
				tx := transaction.RegisterNode{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						m.Data = tx.NodeAddress.Bytes()
					},
				)

			case transaction.UnregisterNodeTx:
				tx := transaction.UnregisterNode{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						// nothing
					},
				)

			case transaction.TransferAssetTx:
				tx := transaction.TransferAsset{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						switch tx.Token {
						case mint.TokenMNT:
							m.AmountMNT = amount.FromAmount(tx.Amount)
							// stat
							blockModel.TotalMNT.Value.Add(blockModel.TotalMNT.Value, tx.Amount.Value)
						case mint.TokenGOLD:
							m.AmountGOLD = amount.FromAmount(tx.Amount)
							// stat
							blockModel.TotalGOLD.Value.Add(blockModel.TotalGOLD.Value, tx.Amount.Value)
						}
					},
				)

			case transaction.SetWalletTagTx:
				tx := transaction.SetWalletTag{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						m.Data = []byte(tx.Tag.String())
					},
				)

			case transaction.UnsetWalletTagTx:
				tx := transaction.UnsetWalletTag{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.Address
						m.To = &to
						m.Data = []byte(tx.Tag.String())
					},
				)

			case transaction.UserDataTx:
				tx := transaction.UserData{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						m.Data = tx.Data
						// stat
						blockModel.TotalUserData += uint64(len(m.Data))
					},
				)

			case transaction.DistributionFeeTx:
				tx := transaction.DistributionFee{}
				return mkTX(
					&tx, t, d, h, p.pubTX,
					func(m *Transaction) {
						to := tx.OwnerAddress
						m.To = &to
						m.AmountMNT = amount.FromAmount(tx.AmountMNT)
						m.AmountGOLD = amount.FromAmount(tx.AmountGOLD)
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

func mkTX(itx transaction.Transactioner, typ transaction.Code, d *serializer.Deserializer, h *block.Header, pubTX chan<- *Transaction, fillCbk ptxCbk) error {
	ptx, err := itx.Parse(d.Source())
	if err != nil {
		return err
	}
	m := Transaction{
		Digest:     ptx.Digest,
		Block:      big.NewInt(0).Set(h.BlockID),
		Type:       typ,
		Nonce:      ptx.Nonce,
		From:       ptx.From,
		To:         nil,
		AmountMNT:  amount.New(),
		AmountGOLD: amount.New(),
		Timestamp:  mint.StampToTime(h.Timestamp),
		Data:       nil,
	}
	fillCbk(&m)
	pubTX <- &m
	return nil
}
