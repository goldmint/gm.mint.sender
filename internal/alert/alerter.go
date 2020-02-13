package alert

import "time"

var (
	// NullAlerter is blackhole alerter
	NullAlerter         = &Null{}
	_           Alerter = NullAlerter
)

// Alerter sends messages to the administrator
type Alerter interface {
	Info(f string, arg ...interface{}) error
	Warn(f string, arg ...interface{}) error
	Error(f string, arg ...interface{}) error
	LimitWarn(max time.Duration, f string, arg ...interface{}) error
	LimitInfo(max time.Duration, f string, arg ...interface{}) error
	LimitError(max time.Duration, f string, arg ...interface{}) error
	LimitTagWarn(max time.Duration, tag string, f string, arg ...interface{}) error
	LimitTagInfo(max time.Duration, tag string, f string, arg ...interface{}) error
	LimitTagError(max time.Duration, tag string, f string, arg ...interface{}) error
}

// ---

// Null is a blackhole
type Null struct {
}

// Info implementation
func (a *Null) Info(_ string, _ ...interface{}) error {
	return nil
}

// Warn implementation
func (a *Null) Warn(_ string, _ ...interface{}) error {
	return nil
}

// Error implementation
func (a *Null) Error(_ string, _ ...interface{}) error {
	return nil
}

// LimitInfo implementation
func (a *Null) LimitInfo(_ time.Duration, _ string, _ ...interface{}) error {
	return nil
}

// LimitWarn implementation
func (a *Null) LimitWarn(_ time.Duration, _ string, _ ...interface{}) error {
	return nil
}

// LimitError implementation
func (a *Null) LimitError(_ time.Duration, _ string, _ ...interface{}) error {
	return nil
}

// LimitTagInfo implementation
func (a *Null) LimitTagInfo(_ time.Duration, _, _ string, _ ...interface{}) error {
	return nil
}

// LimitTagWarn implementation
func (a *Null) LimitTagWarn(_ time.Duration, _, _ string, _ ...interface{}) error {
	return nil
}

// LimitTagError implementation
func (a *Null) LimitTagError(_ time.Duration, _, _ string, _ ...interface{}) error {
	return nil
}
