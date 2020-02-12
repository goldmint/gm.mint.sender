package rpc

import (
	"encoding/json"
)

// Event is an RPC event (server-side initiated message)
type Event struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// GetMethod impl
func (e *Event) GetMethod() string {
	return e.Method
}

func (e *Event) isIncomingMessage() {}
