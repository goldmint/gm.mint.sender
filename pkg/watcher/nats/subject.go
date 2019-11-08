package watcher

//go:generate protoc --go_out=. request.proto
//go:generate protoc --go_out=. event.proto

// Subject getter
func (m AddRemove) Subject() string { return "mintsender.watcher.watch" }

// Subject getter
func (m Refill) Subject() string { return "mintsender.watcher.refill" }
