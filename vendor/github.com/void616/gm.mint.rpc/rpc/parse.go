package rpc

import (
	"encoding/json"
	"errors"
)

// ParseMessage tries to parse message into a proper structure
func ParseMessage(msg []byte) (IncomingMessage, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(msg, &root); err != nil {
		return nil, err
	}

	if _, ok := root["method"]; !ok {
		return nil, errors.New("method field not found")
	}

	_, hasID := root["id"]
	_, hasParams := root["params"]
	_, hasResult := root["result"]
	_, hasError := root["error"]

	var message IncomingMessage

	switch {
	// event
	case !hasID && hasParams:
		message = &Event{}
	// error
	case hasID && hasError:
		message = &Error{}
	// result
	case hasID && hasResult:
		message = &Result{}
	default:
		return nil, errors.New("unknown message type")
	}

	// parse into struct
	if err := json.Unmarshal(msg, message); err != nil {
		return nil, err
	}
	return message, nil
}
