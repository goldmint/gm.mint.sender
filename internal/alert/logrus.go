package alert

import (
	"time"

	"github.com/sirupsen/logrus"
)

var (
	_ Alerter = &LogrusAlerter{}
)

// LogrusAlerter sends alerts to logrus logger
type LogrusAlerter struct {
	logger  *logrus.Entry
	limiter *timeLimiter
}

// NewLogrus instance
func NewLogrus(logger *logrus.Entry) *LogrusAlerter {
	return &LogrusAlerter{
		logger:  logger,
		limiter: newTimeLimiter(),
	}
}

// Info implementation
func (a *LogrusAlerter) Info(f string, arg ...interface{}) error {
	a.logger.Infof(f, arg...)
	return nil
}

// Warn implementation
func (a *LogrusAlerter) Warn(f string, arg ...interface{}) error {
	a.logger.Warnf(f, arg...)
	return nil
}

// Error implementation
func (a *LogrusAlerter) Error(f string, arg ...interface{}) error {
	a.logger.Errorf(f, arg...)
	return nil
}

// LimitInfo implementation
func (a *LogrusAlerter) LimitInfo(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Info(f, arg...)
	}
	return nil
}

// LimitWarn implementation
func (a *LogrusAlerter) LimitWarn(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Warn(f, arg...)
	}
	return nil
}

// LimitError implementation
func (a *LogrusAlerter) LimitError(max time.Duration, f string, arg ...interface{}) error {
	if a.limiter.limit(max, "") {
		return a.Error(f, arg...)
	}
	return nil
}

// LimitTagInfo implementation
func (a *LogrusAlerter) LimitTagInfo(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Info(f, arg...)
	}
	return nil
}

// LimitTagWarn implementation
func (a *LogrusAlerter) LimitTagWarn(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Warn(f, arg...)
	}
	return nil
}

// LimitTagError implementation
func (a *LogrusAlerter) LimitTagError(max time.Duration, tag, f string, arg ...interface{}) error {
	if a.limiter.limit(max, tag) {
		return a.Error(f, arg...)
	}
	return nil
}
