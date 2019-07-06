// Package health serves health checks over HTTP at /health endpoint.
package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Listener is an http health check listener
type Listener struct {
	logger logger
	port   int
	donech chan struct{}
	srv    *http.Server
}

// New create a new http health check listener
func New(log logger, port int) *Listener {
	return &Listener{
		logger: log,
		port:   port,
		donech: make(chan struct{}),
		srv:    nil,
	}
}

func (h *Listener) healthCheckReply(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "ok\n"); err != nil {
		h.logger.Errorf("Failed to reply to http healtcheck from %s: %s\n", r.RemoteAddr, err)
	}
}

// Start exposes an http healthcheck handler
func (h *Listener) Start() *Listener {
	if h.port == 0 {
		return h
	}

	h.logger.Infof("Starting http healtcheck handler")

	h.srv = &http.Server{Addr: fmt.Sprintf(":%d", h.port)}

	http.HandleFunc("/health", h.healthCheckReply)

	go func() {
		defer close(h.donech)
		err := h.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			h.logger.Errorf("healthcheck server failed: %v", err)
		}
	}()

	return h
}

// Stop halts the http health check handler
func (h *Listener) Stop() {
	if h.srv == nil {
		return
	}

	h.logger.Infof("Stopping http healtcheck handler")

	err := h.srv.Shutdown(context.TODO())
	if err != nil {
		h.logger.Errorf("failed to stop http healtcheck handler: %v", err)
	}

	<-h.donech
}
