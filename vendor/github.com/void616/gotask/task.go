package gotask

import (
	"sync/atomic"
)

// Recover is a callback function template that will be called on routine panic.
type Recover func(interface{})

// Task is a named routine that does something in parallel.
type Task struct {
	tag         string
	routine     interface{}
	routineArgs []interface{}
	running     *int32
	recover     Recover
	logger      Logger
}

// NewTask instance. Arg `routine` should be an entity of type:
// func(),
// func(*Token),
// func(...interface{}),
// func(*Token, ...interface{}).
// Otherwise `ErrInvalidRoutine` will be returned.
func NewTask(tag string, routine interface{}, args ...interface{}) (*Task, error) {
	switch routine.(type) {
	case func():
	case func(*Token):
	case func(...interface{}):
	case func(*Token, ...interface{}):
	default:
		return nil, ErrInvalidRoutine
	}
	return &Task{
		tag:         tag,
		routine:     routine,
		routineArgs: args,
		running:     new(int32),
	}, nil
}

// Tag gets Task's tag.
func (t *Task) Tag() string {
	return t.tag
}

// Log sets a Logger. If the Task is running, `ErrInvalidState` error will be returned.
func (t *Task) Log(l Logger) *Task {
	t.logger = l
	return t
}

// Recover specifies a callback function that will be called in case of routine's panic.
func (t *Task) Recover(r Recover) *Task {
	t.recover = r
	return t
}

// Run starts routine in a goroutine and returns `*Token` to control lifetime and `*Waiter` to wait routine stop.
// If the Task is already running, `ErrInvalidState` error will be returned.
func (t *Task) Run() (*Token, *Waiter, error) {
	if !atomic.CompareAndSwapInt32(t.running, 0, 1) {
		return nil, nil, ErrInvalidState
	}

	t.print("Task", t.tag, "running")

	token := newToken(t.tag, t.logger)
	waiter := newWaiter(t.tag, t.logger)

	go func(tsk *Task, tkn *Token, wtr *Waiter) {
		defer func() {
			tkn.Stop()
			wtr.onStopped()
			tsk.print("Task", tsk.tag, "stopped")
			atomic.StoreInt32(tsk.running, 0)
			if r := recover(); r != nil {
				if tsk.recover != nil {
					tsk.recover(r)
				} else {
					panic(r)
				}
			}
		}()
		switch r := tsk.routine.(type) {
		case func():
			r()
		case func(*Token):
			r(tkn)
		case func(...interface{}):
			r(tsk.routineArgs...)
		case func(*Token, ...interface{}):
			r(tkn, tsk.routineArgs...)
		default:
			panic("not implemented")
		}
	}(t, token, waiter)

	return token, waiter, nil
}

func (t *Task) print(args ...interface{}) {
	if t.logger != nil {
		t.logger.Log(args...)
	}
}
