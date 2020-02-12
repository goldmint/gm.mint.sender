package block

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/void616/gm.mint/transaction"
	"golang.org/x/crypto/sha3"

	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint/serializer"
)

// Header data
type Header struct {
	// Version of the blockchain
	Version uint16
	// PrevBlockDigest
	PrevBlockDigest mint.Digest
	// ConsensusRound
	ConsensusRound uint16
	// MerkleRoot
	MerkleRoot mint.Digest
	// Timestamp of the block
	Timestamp uint64
	// TransactionsCount in the block
	TransactionsCount uint16
	// BlockID
	BlockID *big.Int
	// SignersCount
	SignersCount uint16
	// Signers list
	Signers []Signer
	// Digest (header)
	Digest mint.Digest
}

// Signer data
type Signer struct {
	// PublicKey
	PublicKey mint.PublicKey
	// Signature
	Signature mint.Signature
}

// CbkHeader for parsed header
type CbkHeader func(*Header) error

// CbkTransaction for parsed transaction
type CbkTransaction func(transaction.Code, *serializer.Deserializer, *Header) error

// ---

// Parse block
func Parse(r io.Reader, cbkHeader CbkHeader, cbkTransaction CbkTransaction) error {
	d := serializer.NewStreamDeserializer(r)

	// read header data into buffer to get it's digest later
	headerData := bytes.NewBuffer(nil)
	hd := serializer.NewStreamDeserializer(io.TeeReader(r, headerData))

	// read header
	header := &Header{}
	header.Version = hd.GetUint16()         // version
	header.PrevBlockDigest = hd.GetDigest() // previous block digest
	header.ConsensusRound = hd.GetUint16()  // consensus round
	header.MerkleRoot = hd.GetDigest()      // merkle root

	// we should provide timestamp length (4 bytes, uint32, ) to calculate header digest (kinda bug in node's code)
	headerData.WriteByte(8)
	headerData.WriteByte(0)
	headerData.WriteByte(0)
	headerData.WriteByte(0)

	// continue to read header
	header.Timestamp = hd.GetUint64()         // time
	header.TransactionsCount = hd.GetUint16() // transactions
	header.BlockID = hd.GetUint256()          // block
	if err := hd.Error(); err != nil {
		return err
	}

	// calc header digest
	{
		hasher := sha3.New256()
		if _, err := hasher.Write(headerData.Bytes()); err != nil {
			return err
		}
		copy(header.Digest[:], hasher.Sum(nil))
	}

	// continue to read header
	header.SignersCount = d.GetUint16() // signers
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

		code := d.GetUint16() // code
		if err := d.Error(); err != nil {
			return err
		}

		// check the code
		if !transaction.ValidCode(code) {
			return fmt.Errorf("unknown transaction code %v at index %v", code, i)
		}
		txCode := transaction.Code(code)

		// parse transaction outside
		if err := cbkTransaction(txCode, d, header); err != nil {
			return err
		}
		if err := d.Error(); err != nil {
			return err
		}
	}

	return nil
}
