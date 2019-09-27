package model

import (
	"regexp"
)

// ServiceNameRex is unique service name constrain
var ServiceNameRex = regexp.MustCompile("^[a-zA-Z0-9-_]+$")
