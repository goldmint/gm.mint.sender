package rpc

import "encoding/json"

// Result is successful RPC response for some RPC request
type Result struct {
	Method string          `json:"method"`
	ID     uint32          `json:"id"`
	Result json.RawMessage `json:"result"`
}

// GetMethod impl
func (r *Result) GetMethod() string {
	return r.Method
}

// GetID impl
func (r *Result) GetID() uint32 {
	return r.ID
}

// Parse parses inner result structure
func (r *Result) Parse(v interface{}) error {
	return json.Unmarshal(r.Result, v)
}

func (r *Result) isIncomingMessage() {}

func (r *Result) isResponse() {}
