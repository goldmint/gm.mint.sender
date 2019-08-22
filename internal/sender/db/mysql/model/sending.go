package model

// SendingStatus enum
type SendingStatus uint8

const (
	// SendingEnqueued means sending just enqueued
	SendingEnqueued SendingStatus = iota
	// SendingPosted means sender has sent a transaction
	SendingPosted
	// SendingConfirmed means sent transaction is confirmed (shown in some block)
	SendingConfirmed
	// SendingFailed means failure
	SendingFailed
)

// Sending model
type Sending struct {
	Base
	ID          uint64        `gorm:"PRIMARY_KEY;AUTO_INCREMENT:true"`
	Status      SendingStatus `gorm:"NOT NULL"`
	Notified    uint8         `gorm:"NOT NULL"`
	To          []byte        `gorm:"SIZE:32;NOT NULL"`
	Token       uint16        `gorm:"NOT NULL"`
	Amount      string        `gorm:"NOT NULL" sql:"TYPE:decimal(30,18)"`
	Sender      []byte        `gorm:"SIZE:32"`
	SenderNonce uint64        `gorm:""`
	Digest      []byte        `gorm:"SIZE:32"`
	SentAtBlock uint64        `gorm:"SIZE:32"`
	Block       uint64        `gorm:"SIZE:32"`
	RequestID   string        `gorm:"SIZE:64;NOT NULL"`
}
