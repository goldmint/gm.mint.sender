package gotask

import (
	"sync"
)

// Group is a group of Tasks.
type Group struct {
	tag     string
	tasks   []*Task
	tokens  map[*Task]*Token
	waiters map[*Task]*Waiter
	mon     sync.Mutex
	running bool
	logger  Logger
}

// NewGroup instance.
func NewGroup(tag string) *Group {
	return &Group{
		tag:    tag,
		logger: nil,
	}
}

// Tag gets Group's tag.
func (g *Group) Tag() string {
	return g.tag
}

// Log sets a Logger. If the Group is running, `ErrInvalidState` error will be returned.
func (g *Group) Log(l Logger) error {
	g.mon.Lock()
	defer g.mon.Unlock()

	if g.running {
		return ErrInvalidState
	}
	g.logger = l
	return nil
}

// Add adds a Task. If the Group is running, `ErrInvalidState` error will be returned.
func (g *Group) Add(t *Task) error {
	g.mon.Lock()
	defer g.mon.Unlock()

	if g.running || t == nil {
		return ErrInvalidState
	}
	g.tasks = append(g.tasks, t)
	g.print("Group", g.tag, "|", "Task", t.Tag(), "added")
	return nil
}

// Run starts Group's Tasks.
// If the Group is running, `ErrInvalidState` error will be returned.
// Run will ignore any Task that is already running (i.e. Task.Run() return error).
func (g *Group) Run() error {
	g.mon.Lock()
	defer g.mon.Unlock()

	if g.running {
		return ErrInvalidState
	}

	if len(g.tasks) == 0 {
		return ErrInvalidState
	}

	g.print("Group", g.tag, "running")

	tokens := make(map[*Task]*Token)
	waiters := make(map[*Task]*Waiter)

	for _, task := range g.tasks {
		t, w, err := task.Run()
		if err != nil {
			g.print("Group", g.tag, "|", "Task", task.Tag(), "skipped to run")
			continue
		}
		tokens[task] = t
		waiters[task] = w
	}

	g.tokens = tokens
	g.waiters = waiters
	g.running = true
	return nil
}

// Stop requires Group's Tasks to stop. Does nothing if the Group is not running.
func (g *Group) Stop() {
	g.mon.Lock()
	defer g.mon.Unlock()

	if !g.running {
		return
	}
	g.print("Group", g.tag, "required to stop")
	for _, token := range g.tokens {
		token.Stop()
	}
}

// Wait awaits Group's Tasks stop. Does nothing if the Group is not running.
func (g *Group) Wait() {
	g.mon.Lock()

	if !g.running {
		g.mon.Unlock()
		return
	}
	g.print("Group", g.tag, "waiting")
	g.mon.Unlock()

	for _, waiter := range g.waiters {
		waiter.Wait()
		g.print("Group", g.tag, "|", "Task", waiter.Tag(), "stopped")
	}

	g.mon.Lock()
	g.tokens = nil
	g.waiters = nil
	g.running = false
	g.print("Group", g.tag, "stopped")
	g.mon.Unlock()
}

// TokenOf returns specific Task Token.
// If the Task wasn't started by the Group or Group isn't running, `ErrInvalidState` will be returned.
func (g *Group) TokenOf(t *Task) (*Token, error) {
	g.mon.Lock()
	defer g.mon.Unlock()

	if !g.running {
		return nil, ErrInvalidState
	}

	token, ok := g.tokens[t]
	if !ok {
		return nil, ErrInvalidState
	}
	return token, nil
}

// WaiterOf returns specific Task Waiter.
// If the Task wasn't started by the Group or Group isn't running, `ErrInvalidState` will be returned.
func (g *Group) WaiterOf(t *Task) (*Waiter, error) {
	g.mon.Lock()
	defer g.mon.Unlock()

	if !g.running {
		return nil, ErrInvalidState
	}

	waiter, ok := g.waiters[t]
	if !ok {
		return nil, ErrInvalidState
	}
	return waiter, nil
}

func (g *Group) print(args ...interface{}) {
	if g.logger != nil {
		g.logger.Log(args...)
	}
}
