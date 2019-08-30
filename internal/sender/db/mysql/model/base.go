package model

import (
	"time"
	"unicode/utf8"
)

// Base DB entity
type Base struct {
	CreatedAt time.Time `gorm:"NOT NULL;DEFAULT:current_timestamp"`
	UpdatedAt time.Time `gorm:"NOT NULL;DEFAULT:current_timestamp"`
}

// LimitStringField crops string
func LimitStringField(s string, maxRunes uint) string {
	if uint(utf8.RuneCountInString(s)) > maxRunes {
		charz := make([]rune, maxRunes)
		copy(charz, []rune(s))
		return string(charz)
	}
	return s
}
