package request

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/void616/gm.mint/amount"
)

// Balance is common balance struct (blockchain balance, wallet balance)
type Balance struct {
	Gold *amount.Amount `json:"gold"`
	Mnt  *amount.Amount `json:"mnt"`
}

func (b *Balance) checkValues() {
	if b.Gold == nil {
		b.Gold = amount.New()
	}
	if b.Mnt == nil {
		b.Mnt = amount.New()
	}
}

// ---

// BigInt is big.Int wrapper
type BigInt struct {
	*big.Int
}

// MarshalJSON impl
func (bi *BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(bi.Text(10))
}

// UnmarshalJSON impl.
func (bi *BigInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	tmp := new(big.Int)
	_, ok := tmp.SetString(s, 10)
	if !ok {
		return fmt.Errorf("failed to parse big int from `%v`", s)
	}
	bi.Int = tmp
	return nil
}

// ---

// ByteArray is byte array wrapper
type ByteArray []byte

// MarshalJSON impl
func (ba *ByteArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(*ba))
}

// UnmarshalJSON impl.
func (ba *ByteArray) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	bb, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("failed to parse byte array from `%v`: %v", s, err)
	}
	*ba = bb
	return nil
}
