package types

// SendingTransport is a type of transcport of the API, i.e. HTTP, Nats etc.
type SendingTransport uint8

const (
	// SendingNats is Nats transport
	SendingNats SendingTransport = iota + 1
	// SendingHTTP is HTTP transport
	SendingHTTP
)

// SendingStatus enum
type SendingStatus uint8

const (
	// SendingEnqueued means sending just enqueued
	SendingEnqueued SendingStatus = 0
	// SendingPosted means sender has sent a transaction
	SendingPosted SendingStatus = 1
	// SendingConfirmed means sent transaction is confirmed (shown in some block)
	SendingConfirmed SendingStatus = 2
	// SendingFailed means failure
	SendingFailed SendingStatus = 3
)
