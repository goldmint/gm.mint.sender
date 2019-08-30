package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/void616/gotask"
)

// Service is metrics service that serves Prometheus /metrics endpoint
type Service struct {
	logger *logrus.Entry
	port   uint16
}

// New instance
func New(port uint16, logger *logrus.Entry) *Service {
	return &Service{
		port:   port,
		logger: logger,
	}
}

// Task loop
func (s *Service) Task(token *gotask.Token) {
	wg := sync.WaitGroup{}

	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.port),
		Handler: promhttp.Handler(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Infof("Serving metrics on port %v", s.port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Failed to setup server")
		}
	}()

	for !token.Stopped() {
		token.Sleep(time.Second)
	}

	if err := server.Shutdown(nil); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown server")
	}

	wg.Wait()
}
