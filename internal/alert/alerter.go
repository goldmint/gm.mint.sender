package alert

import "time"

var (
	// NullAlerter is blackhole alerter
	NullAlerter         = &Null{}
	_           Alerter = NullAlerter
)

// Alerter sends messages to the administrator
type Alerter interface {
	Info(f string, arg ...interface{})
	Warn(f string, arg ...interface{})
	Error(f string, arg ...interface{})
	LimitWarn(max time.Duration, f string, arg ...interface{})
	LimitInfo(max time.Duration, f string, arg ...interface{})
	LimitError(max time.Duration, f string, arg ...interface{})
}

// ---

// Null is a blackhole
type Null struct {
}

// Info implementation
func (a *Null) Info(_ string, _ ...interface{}) {
}

// Warn implementation
func (a *Null) Warn(_ string, _ ...interface{}) {
}

// Error implementation
func (a *Null) Error(_ string, _ ...interface{}) {
}

// LimitInfo implementation
func (a *Null) LimitInfo(_ time.Duration, _ string, _ ...interface{}) {
}

// LimitWarn implementation
func (a *Null) LimitWarn(_ time.Duration, _ string, _ ...interface{}) {
}

// LimitError implementation
func (a *Null) LimitError(_ time.Duration, _ string, _ ...interface{}) {
}
