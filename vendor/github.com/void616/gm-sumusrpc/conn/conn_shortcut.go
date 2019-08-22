package conn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// SendThenReceiveMessage shortcut
func (r *Conn) SendThenReceiveMessage(msg []byte, id string) ([]byte, error) {
	if r.Closing() {
		return nil, fmt.Errorf("connection is closed")
	}

	ch := r.Subscribe()
	if ch == nil {
		return nil, fmt.Errorf("connection is closed?")
	}
	defer r.Unsubscribe()

	// send data
	if err := r.sendMessage(msg); err != nil {
		return nil, err
	}

	for {
		if r.Closing() {
			return nil, fmt.Errorf("connection is closed")
		}

		var data []byte

		select {
		case evt := <-ch:
			if evt.Error != nil {
				return nil, fmt.Errorf("message receiving timeout")
			}
			data = evt.Message
		case <-time.After(r.recvTimeout):
			return nil, fmt.Errorf("message receiving timeout")
		}

		// get response as json
		var model map[string]*json.RawMessage
		err := json.Unmarshal(data, &model)
		if err != nil {
			return nil, err
		}

		// check id field
		jid, ok := model["id"]
		if !ok {
			return nil, fmt.Errorf("failed to find `id` field")
		}

		// get id field as string
		sid := ""
		err = json.Unmarshal(*jid, &sid)
		if err != nil {
			return nil, fmt.Errorf("failed to parse `id` field")
		}

		// test id
		if sid == id {
			return data, nil
		}
	}
}

// Heartbeat shortcut
func (r *Conn) Heartbeat() error {
	if r.Closing() {
		return fmt.Errorf("connection is closed")
	}

	// send
	b, err := r.SendThenReceiveMessage(
		[]byte(`{"id":"get-blockchain-state"}`),
		"get-blockchain-state",
	)
	if err != nil {
		return err
	}

	// response check
	if !bytes.Contains(b, []byte(`"block_count"`)) {
		return fmt.Errorf("unexpected response")
	}
	return nil
}
