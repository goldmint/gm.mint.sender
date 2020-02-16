package http

// WatchRequest is /watch request model
type WatchRequest struct {
	Service    string   `json:"service"`     // Service name (to differentiate multiple requestors): 1..64
	PublicKeys []string `json:"public_keys"` // Destination wallet address in Base58
	Callback   string   `json:"callback"`    // Callback for notification: 1..256
}

// UnwatchRequest is /unwatch request model
type UnwatchRequest struct {
	Service    string   `json:"service"`     // Service name (to differentiate multiple requestors): 1..64
	PublicKeys []string `json:"public_keys"` // Destination wallet address in Base58
}

// RefillEvent is notification model
type RefillEvent struct {
	Service     string `json:"service"`     // Service name (to differentiate multiple requestors): 1..64
	PublicKey   string `json:"public_key"`  // Destination (watching) wallet address in Base58
	From        string `json:"from"`        // Source wallet address in Base58
	Token       string `json:"token"`       // GOLD or MNT
	Amount      string `json:"amount"`      // Token amount in major units: 1.234 (18 decimal places)
	Transaction string `json:"transaction"` // Digest of the refilling tx in Base58
}
