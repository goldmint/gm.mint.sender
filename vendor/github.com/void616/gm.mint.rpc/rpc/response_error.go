package rpc

import "fmt"

// Error is an error RPC response for some RPC request
type Error struct {
	Method string `json:"method"`
	ID     uint32 `json:"id"`
	Error  struct {
		Code   ErrorCode    `json:"code"`
		Desc   string       `json:"desc"`
		Reason *ErrorReason `json:"reason,omitempty"`
	} `json:"error"`
}

// ErrorReason extends Error
type ErrorReason struct {
	Code ErrorCode `json:"code"`
	Desc string    `json:"desc"`
}

// GetMethod impl
func (e *Error) GetMethod() string {
	return e.Method
}

// GetID impl
func (e *Error) GetID() uint32 {
	return e.ID
}

// GetReason tries to get reason
func (e *Error) GetReason() (code ErrorCode, desc string, ok bool) {
	if e.Error.Reason != nil {
		return e.Error.Reason.Code, e.Error.Reason.Desc, true
	}
	return EUnclassified, "", false
}

// Err makes Go error
func (e *Error) Err() error {
	code, desc, ok := e.GetReason()
	if ok {
		return fmt.Errorf("rpc reason code %d (%v): %v", code, code.String(), desc)
	}
	return fmt.Errorf("rpc code %d (%v): %v", e.Error.Code, e.Error.Code.String(), e.Error.Desc)
}

func (e *Error) isIncomingMessage() {}

func (e *Error) isResponse() {}
