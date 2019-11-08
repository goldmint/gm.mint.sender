package sender

//go:generate protoc --go_out=. request.proto
//go:generate protoc --go_out=. event.proto

// Subject getter
func (m Send) Subject() string { return "mintsender.sender.send" }

// Subject getter
func (m Sent) Subject() string { return "mintsender.sender.sent" }
