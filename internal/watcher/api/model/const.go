package model

import (
	"net/url"
	"regexp"
	"strings"
)

// ServiceNameRex is service name pattern
var ServiceNameRex = regexp.MustCompile("^[a-z0-9_]{1,64}$")

// ValidCallback checker
func ValidCallback(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.IsAbs() && strings.HasPrefix(strings.ToLower(u.Scheme), "http")
}
