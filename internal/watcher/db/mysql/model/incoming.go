package model

import "time"

// Incoming model
type Incoming struct {
	Base
	To        []byte    `gorm:"SIZE:32;NOT NULL"`
	From      []byte    `gorm:"SIZE:32;NOT NULL"`
	Amount    string    `gorm:"NOT NULL" sql:"TYPE:decimal(30,18)"`
	Token     uint16    `gorm:"NOT NULL"`
	Digest    []byte    `gorm:"PRIMARY_KEY;SIZE:32;NOT NULL"`
	Block     []byte    `gorm:"SIZE:32;NOT NULL"`
	Timestamp time.Time `gorm:"NOT NULL;DEFAULT:current_timestamp"`
	Sent      bool      `gorm:"NOT NULL"`
}
