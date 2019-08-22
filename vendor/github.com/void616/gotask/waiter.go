package gotask

import (
	"time"
)

// Waiter allows to await a Task.
type Waiter struct {
	tag     string
	stopped chan struct{}
	logger  Logger
}

// newWaiter instance.
func newWaiter(tag string, logger Logger) *Waiter {
	return &Waiter{
		tag:     tag,
		stopped: make(chan struct{}),
		logger:  logger,
	}
}

// onStopped called on related Task's stop.
func (w *Waiter) onStopped() {
	close(w.stopped)
}

// Tag gets related Task's tag.
func (w *Waiter) Tag() string {
	return w.tag
}

// Wait awaits Task stop.
func (w *Waiter) Wait() {
	<-w.stopped
}

// WaitTimeout awaits Task stop specified amount of time.
// Returns `true` if the Task is stopped.
func (w *Waiter) WaitTimeout(d time.Duration) bool {
	select {
	case <-w.stopped:
		return true
	case <-time.After(d):
		return false
	}
}

func (w *Waiter) print(args ...interface{}) {
	if w.logger != nil {
		w.logger.Log(args...)
	}
}
