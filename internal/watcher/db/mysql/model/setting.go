package model

// Setting model
type Setting struct {
	Base
	Key   string `gorm:"PRIMARY_KEY;SIZE:128;NOT NULL"`
	Value string `gorm:"SIZE:1024;NOT NULL"`
}
