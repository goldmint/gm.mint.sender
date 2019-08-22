package block

import (
	"fmt"
	"io"
	"math/big"

	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/serializer"
)

// Header data
type Header struct {
	// Version of the blockchain
	Version uint16
	// PrevBlockDigest
	PrevBlockDigest sumuslib.Digest
	// MerkleRoot
	MerkleRoot sumuslib.Digest
	// Timestamp of the block
	Timestamp uint64
	// TransactionsCount in the block
	TransactionsCount uint16
	// BlockNumber
	BlockNumber *big.Int
	// SignersCount
	SignersCount uint16
	// Signers list
	Signers []Signer
}

// Signer data
type Signer struct {
	// PublicKey
	PublicKey sumuslib.PublicKey
	// Signature
	Signature sumuslib.Signature
}

// CbkHeader for parsed header
type CbkHeader func(*Header) error

// CbkTransaction for parsed transaction
type CbkTransaction func(sumuslib.Transaction, *serializer.Deserializer, *Header) error

// ---

// Parse block
func Parse(r io.Reader, cbkHeader CbkHeader, cbkTransaction CbkTransaction) error {

	d := serializer.NewStreamDeserializer(r)

	// read header
	header := &Header{}
	header.Version = d.GetUint16()           // version
	header.PrevBlockDigest = d.GetDigest()   // previous block digest
	header.MerkleRoot = d.GetDigest()        // merkle root
	header.Timestamp = d.GetUint64()         // time
	header.TransactionsCount = d.GetUint16() // transactions
	header.BlockNumber = d.GetUint256()      // block
	header.SignersCount = d.GetUint16()      // signers
	if err := d.Error(); err != nil {
		return err
	}

	// read signers list
	header.Signers = make([]Signer, header.SignersCount)
	for i := uint16(0); i < header.SignersCount; i++ {

		sig := Signer{}
		sig.PublicKey = d.GetPublicKey() // address
		sig.Signature = d.GetSignature() // signature

		if err := d.Error(); err != nil {
			return err
		}
		header.Signers[i] = sig
	}

	// callback
	if err := cbkHeader(header); err != nil {
		return err
	}

	// read transactions
	for i := uint16(0); i < header.TransactionsCount; i++ {

		txCode := d.GetUint16() // code
		if err := d.Error(); err != nil {
			return err
		}

		// check the code
		if !sumuslib.ValidTransaction(txCode) {
			return fmt.Errorf("Unknown transaction with code `%v` with index %v", txCode, i)
		}
		txType := sumuslib.Transaction(txCode)

		// parse transaction outside
		if err := cbkTransaction(txType, d, header); err != nil {
			return err
		}
		if err := d.Error(); err != nil {
			return err
		}
	}

	return nil
}
