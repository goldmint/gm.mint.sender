package types

// Service model
type Service struct {
	ID          uint64
	Name        string
	Transport   ServiceTransport
	CallbackURL string
}

// ServiceTransport is a type of transport of the API, i.e. HTTP, Nats etc.
type ServiceTransport uint8

const (
	// ServiceNats is Nats transport
	ServiceNats ServiceTransport = iota + 1
	// ServiceHTTP is HTTP transport
	ServiceHTTP
)
