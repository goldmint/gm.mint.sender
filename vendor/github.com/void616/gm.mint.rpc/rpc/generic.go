package rpc

// IncomingMessage is an incoming message: Response or Event
type IncomingMessage interface {
	isIncomingMessage()
	GetMethod() string
}

// Response is an incoming response: Result or Error
type Response interface {
	IncomingMessage
	isResponse()
	GetID() uint32
}
