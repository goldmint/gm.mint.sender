package conn

import (
	"log"
	"time"
)

// DefaultOptions contain nothing special
var DefaultOptions = Options{
	ID:          "noid",
	ConnTimeout: time.Second * 10,
}

// Options are options
type Options struct {
	Logger      func(id, msg string)
	ID          string
	ConnTimeout time.Duration
}

// WithStdLogger sets standard logger
func (o Options) WithStdLogger() Options {
	ret := o
	ret.Logger = func(id, msg string) {
		log.Printf("conn %v: %v", id, msg)
	}
	return ret
}
