package blockparser

import (
	"math/big"
	"time"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gm.mint/transaction"
)

// Block is a block header data
type Block struct {
	// Block is ID
	Block *big.Int
	// PrevDigest is previous block digest
	PrevDigest mint.Digest
	// MerkleRoot of the block
	MerkleRoot mint.Digest
	// TransactionsCount in the block
	TransactionsCount uint16
	// Signers is an array of signers' public keys
	Signers []mint.PublicKey
	// TotalMNT transferred in the block
	TotalMNT *amount.Amount
	// TotalGOLD transferred in the block
	TotalGOLD *amount.Amount
	// FeeMNT is an amount of collected fee in MNT (owner wallet)
	FeeMNT *amount.Amount
	// FeeGOLD is an amount of collected fee in GOLD (owner wallet)
	FeeGOLD *amount.Amount
	// TotalUserData is an amount of bytes transferred in UserData transactions
	TotalUserData uint64
	// Timestamp of the block, UTC
	Timestamp time.Time
}

// Transaction is a single transaction data
type Transaction struct {
	// Digest of the tx
	Digest mint.Digest
	// Block is a block where the tx fixed
	Block *big.Int
	// Type of the tx
	Type transaction.Code
	// Nonce is a nonce/ID of the tx
	Nonce uint64
	// From is a sender's public key
	From mint.PublicKey
	// To is a receivers's public key (where applicable) or nil
	To *mint.PublicKey
	// AmountMNT is an amount of MNT transferred
	AmountMNT *amount.Amount
	// AmountGOLD is an amount of GOLD transferred
	AmountGOLD *amount.Amount
	// Timestamp of the tx' block, UTC
	Timestamp time.Time
	// Data is an optional payload bytes
	Data []byte
}
