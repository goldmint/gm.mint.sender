package model

import "time"

// Base DB entity
type Base struct {
	CreatedAt time.Time `gorm:"NOT NULL;DEFAULT:current_timestamp"`
	UpdatedAt time.Time `gorm:"NOT NULL;DEFAULT:current_timestamp"`
}
