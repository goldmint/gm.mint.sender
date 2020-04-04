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
func (a *LogrusAlerter) Info(f string, arg ...interface{}) {
	a.logger.Infof(f, arg...)
}

// Warn implementation
func (a *LogrusAlerter) Warn(f string, arg ...interface{}) {
	a.logger.Warnf(f, arg...)
}

// Error implementation
func (a *LogrusAlerter) Error(f string, arg ...interface{}) {
	a.logger.Errorf(f, arg...)
}

// LimitInfo implementation
func (a *LogrusAlerter) LimitInfo(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Info(f, arg...)
	}
}

// LimitWarn implementation
func (a *LogrusAlerter) LimitWarn(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Warn(f, arg...)
	}
}

// LimitError implementation
func (a *LogrusAlerter) LimitError(max time.Duration, f string, arg ...interface{}) {
	if a.limiter.limit(max, f, arg...) {
		a.Error(f, arg...)
	}
}
