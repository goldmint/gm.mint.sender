package rpc

import (
	"encoding/json"
)

// Request is an RPC request
type Request struct {
	Method string      `json:"method"`
	ID     uint32      `json:"id"`
	Params interface{} `json:"params"`
}

// JSON ...
func (r *Request) JSON() ([]byte, error) {
	// has to provide params object anyway
	if r.Params == nil {
		r.Params = struct{}{}
	}
	return json.Marshal(r)
}
