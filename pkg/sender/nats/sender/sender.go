package sender

//go:generate protoc --go_out=. sender.proto

const (
	// SubjectSend request subject
	SubjectSend = "mintsender.sender.send"
	// SubjectSent event subject
	SubjectSent = "mintsender.sender.sent"
)
