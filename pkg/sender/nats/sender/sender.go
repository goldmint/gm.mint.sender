package sender

//go:generate protoc --go_out=. sender.proto

const (
	// SubjectSend request subject
	SubjectSend = "sender.send"
	// SubjectSent event subject
	SubjectSent = "sender.sent"
)
