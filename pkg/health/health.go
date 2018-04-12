// Package health serves health checks over HTTP at /health endpoint.
package health

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bpineau/katafygio/config"
)

// Listener is an http health check listener
type Listener struct {
	config *config.KfConfig
	donech chan struct{}
	srv    *http.Server
}

// New create a new http health check listener
func New(config *config.KfConfig) *Listener {
	return &Listener{
		config: config,
		donech: make(chan struct{}),
		srv:    nil,
	}
}

func (h *Listener) healthCheckReply(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "ok\n"); err != nil {
		h.config.Logger.Warningf("Failed to reply to http healtcheck from %s: %s\n", r.RemoteAddr, err)
	}
}

// Start exposes an http healthcheck handler
func (h *Listener) Start() *Listener {
	if h.config.HealthPort == 0 {
		return h
	}

	h.config.Logger.Info("Starting http healtcheck handler")

	h.srv = &http.Server{Addr: fmt.Sprintf(":%d", h.config.HealthPort)}

	http.HandleFunc("/health", h.healthCheckReply)

	go func() {
		defer close(h.donech)
		err := h.srv.ListenAndServe()
		if err != nil && err.Error() != "http: Server closed" {
			h.config.Logger.Errorf("healthcheck server failed: %v", err)
		}
	}()

	return h
}

// Stop halts the http health check handler
func (h *Listener) Stop() {
	if h.srv == nil {
		return
	}

	h.config.Logger.Info("Stopping http healtcheck handler")

	err := h.srv.Shutdown(context.TODO())
	if err != nil {
		h.config.Logger.Warningf("failed to stop http healtcheck handler: %v", err)
	}

	<-h.donech
}
