package wallet

//go:generate protoc --go_out=. wallet.proto

const (
	// SubjectWatch request subject
	SubjectWatch = "mintsender.watcher.watch"
	// SubjectRefill event subject
	SubjectRefill = "mintsender.watcher.refill"
)
