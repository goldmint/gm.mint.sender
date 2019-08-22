package model

// Wallet model
type Wallet struct {
	Base
	PublicKey []byte `gorm:"PRIMARY_KEY;SIZE:32;NOT NULL"`
}
