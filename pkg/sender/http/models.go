package http

import (
	"fmt"
)

// SendRequest is /send request model
type SendRequest struct {
	Service           string `json:"service"`            // Service name (to differentiate multiple requestors): 1..64
	ID                string `json:"id"`                 // Unique request ID (within service): 1..64
	PublicKey         string `json:"public_key"`         // Destination wallet address in Base58
	Token             string `json:"token"`              // GOLD or MNT
	Amount            string `json:"amount"`             // Token amount in major units: 1.234 (18 decimal places)
	Callback          string `json:"callback"`           // Callback for notification: 1..256 or empty
	IgnoreApprovement bool   `json:"ignore_approvement"` // Indicates wallet may not be approved (valid only if sender has 'emission' tag)
}

// String implementation
func (sr SendRequest) String() string {
	return fmt.Sprintf("id%v;%v%v;%v", sr.ID, sr.Amount, sr.Token, sr.PublicKey)
}

// ApproveRequest is /send request model
type ApproveRequest struct {
	Service   string `json:"service"`    // Service name (to differentiate multiple requestors): 1..64
	ID        string `json:"id"`         // Unique request ID (within service): 1..64
	PublicKey string `json:"public_key"` // Destination wallet address in Base58
	Callback  string `json:"callback"`   // Callback for notification: 1..256 or empty
}

// String implementation
func (sr ApproveRequest) String() string {
	return fmt.Sprintf("id%v;%v", sr.ID, sr.PublicKey)
}

// SentEvent is notification model
type SentEvent struct {
	Success     bool   `json:"success"`     // Success is true in case of success
	Error       string `json:"error"`       // Error contains error descrition in case of failure
	Service     string `json:"service"`     // Service name (to differentiate multiple requestors): 1..64
	ID          string `json:"id"`          // Unique request ID: 1..64
	PublicKey   string `json:"public_key"`  // Destination wallet address in Base58 (empty on failure)
	Token       string `json:"token"`       // GOLD or MNT (empty on failure)
	Amount      string `json:"amount"`      // Token amount in major units: 1.234 (18 decimal places, empty on failure)
	Transaction string `json:"transaction"` // Transaction digest in Base58 (empty on failure)
}

// ApprovedEvent is notification model
type ApprovedEvent struct {
	Success     bool   `json:"success"`     // Success is true in case of success
	Error       string `json:"error"`       // Error contains error descrition in case of failure
	Service     string `json:"service"`     // Service name (to differentiate multiple requestors): 1..64
	ID          string `json:"id"`          // Unique request ID: 1..64
	PublicKey   string `json:"public_key"`  // Destination wallet address in Base58 (empty on failure)
	Transaction string `json:"transaction"` // Transaction digest in Base58 (empty on failure)
}
