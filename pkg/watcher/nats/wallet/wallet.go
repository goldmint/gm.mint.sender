package wallet

//go:generate protoc --go_out=. wallet.proto

const (
	// SubjectWatch request subject
	SubjectWatch = "watcher.watch"
	// SubjectReceived event subject
	SubjectReceived = "watcher.received"
)
