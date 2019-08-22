package rpc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/void616/gm-sumusrpc/conn"
)

// rpcRequest is a request model
type rpcRequest struct {
	// ID is command
	ID string `json:"id,omitempty"`
	// Params are request params
	Params interface{} `json:"params,omitempty"`
}

// rpcResponse is a response model
type rpcResponse struct {
	// ID is command
	ID string `json:"id,omitempty"`
	// Result is error code
	Result string `json:"result,omitempty"`
	// Text is error message
	Text string `json:"text,omitempty"`
	// Params are response params
	Params interface{} `json:"params,omitempty"`
}

// RawCall sends request `req` to the node via connection `c`, waits for response and deserializes it into `res`.
// In case of transport problems `err` is non-nil and `code` is not set.
func RawCall(c *conn.Conn, command string, req interface{}, res interface{}) (code ErrorCode, err error) {
	code, err = ECUnclassified, nil

	reqModel := rpcRequest{ID: command, Params: req}
	reqBytes, err := json.Marshal(&reqModel)
	if err != nil {
		err = fmt.Errorf("failed to marshal: %v", err.Error())
		return
	}

	// send request and receive response on exact command/ID
	resBytes, err := c.SendThenReceiveMessage(reqBytes, command)
	if err != nil {
		err = fmt.Errorf("failed to make RPC call: %v", err.Error())
		return
	}

	//log.Print("RPC RESPONSE:\n", hex.Dump(resBytes))

	// parse node reponse
	resModel := rpcResponse{Params: res}
	err = json.Unmarshal(resBytes, &resModel)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal: %v", err.Error())
		return
	}

	// parse result code
	resCode, err := strconv.ParseUint(resModel.Result, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse result code: %v", err.Error())
		return
	}

	code = ErrorCode(resCode)
	return
}
