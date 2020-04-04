package alert

import (
	"fmt"
	"sync"
	"time"
)

type timeLimiter struct {
	lock    sync.Mutex
	sources map[string]time.Time
}

func newTimeLimiter() *timeLimiter {
	return &timeLimiter{
		sources: make(map[string]time.Time),
	}
}

func (l *timeLimiter) limit(max time.Duration, f string, arg ...interface{}) bool {
	source := fmt.Sprintf(f, arg...)

	l.lock.Lock()
	defer l.lock.Unlock()

	last, ok := l.sources[source]
	if !ok {
		l.sources[source] = time.Now()
		return true
	}
	if time.Since(last) >= max {
		l.sources[source] = time.Now()
		return true
	}
	return false
}
