package gotask

import (
	"sync/atomic"
	"time"
)

// Token controls a Task's lifetime.
type Token struct {
	tag          string
	stopSignaled *int32
	stop         chan struct{}
	logger       Logger
}

// newToken instance.
func newToken(tag string, logger Logger) *Token {
	return &Token{
		tag:          tag,
		stopSignaled: new(int32),
		stop:         make(chan struct{}),
		logger:       logger,
	}
}

// Tag gets related Task's tag.
func (t *Token) Tag() string {
	return t.tag
}

// Stopped returns `true` if related Task should be stopped.
// In general case this method is used from within a routine.
func (t *Token) Stopped() bool {
	select {
	case <-t.stop:
		return true
	default:
		return false
	}
}

// Sleep sleeps specified amount of time or until related Task is requested to stop.
// Returns `true` if a stop-event received while sleeping.
// In general case this method is used from within a routine.
func (t *Token) Sleep(d time.Duration) bool {
	select {
	case <-t.stop:
		return true
	case <-time.After(d):
		return false
	}
}

// Stop requires related Task to stop.
func (t *Token) Stop() {
	if atomic.CompareAndSwapInt32(t.stopSignaled, 0, 1) {
		t.print("Task", t.tag, "requested to stop")
		close(t.stop)
	}
}

func (t *Token) print(args ...interface{}) {
	if t.logger != nil {
		t.logger.Log(args...)
	}
}
