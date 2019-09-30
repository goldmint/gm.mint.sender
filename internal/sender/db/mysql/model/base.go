package model

import (
	"unicode/utf8"
)

// LimitStringField crops string
func LimitStringField(s string, maxRunes uint) string {
	if uint(utf8.RuneCountInString(s)) > maxRunes {
		charz := make([]rune, maxRunes)
		copy(charz, []rune(s))
		return string(charz)
	}
	return s
}
