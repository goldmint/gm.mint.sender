package http

import (
	"fmt"
	gohttp "net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	mint "github.com/void616/gm.mint"
	"github.com/void616/gm.mint.sender/internal/sender/db/types"
	"github.com/void616/gm.mint/amount"
	"github.com/void616/gotask"
)

// HTTP exposes endpoints via HTTP server
type HTTP struct {
	logger  *logrus.Entry
	api     API
	server  *gohttp.Server
	metrics *Metrics
}

// API provides ability to interact with service API
type API interface {
	EnqueueSending(trans types.SendingTransport, id, service, callbackURL string, to mint.PublicKey, amo *amount.Amount, token mint.Token, ignoreApprovement bool) (dup, success bool)
	EnqueueApprovement(trans types.SendingTransport, id, service, callbackURL string, to mint.PublicKey) (dup, success bool)
}

// New instance
func New(
	port uint,
	api API,
	logger *logrus.Entry,
) (*HTTP, error) {

	var r = mux.NewRouter()
	var server = &gohttp.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        r,
		ReadTimeout:    time.Second * 10,
		WriteTimeout:   time.Second * 10,
		IdleTimeout:    time.Second * 10,
		MaxHeaderBytes: 4096,
	}

	var h = &HTTP{
		logger: logger,
		api:    api,
		server: server,
	}

	r.Path("/send").Methods("POST").HandlerFunc(h.send)
	r.Path("/approve").Methods("POST").HandlerFunc(h.approve)

	logger.Infof("HTTP transport enabled on port %v", port)
	return h, nil
}

// Metrics data
type Metrics struct {
	RequestDuration      *prometheus.HistogramVec
	NotificationDuration prometheus.Histogram
}

// AddMetrics adds metrics counters and should be called before service launch
func (h *HTTP) AddMetrics(m *Metrics) {
	h.metrics = m
}

// Task loop
func (h *HTTP) Task(token *gotask.Token) {

	var wg = sync.WaitGroup{}

	// serve
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := h.server.ListenAndServe(); err != nil && err != gohttp.ErrServerClosed {
			h.logger.WithError(err).Error("Failed to listen and serve")
			token.Stop()
		}
	}()

	// wait
	for !token.Stopped() {
		token.Sleep(time.Millisecond * 500)
	}
	h.server.Shutdown(nil)

	wg.Wait()
}
