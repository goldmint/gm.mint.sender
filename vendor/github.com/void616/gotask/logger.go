package gotask

// Logger logs events for debug
type Logger interface {
	Log(...interface{})
}
