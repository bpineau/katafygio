// Package health serves healthchecks over HTTP at /health endpoint.
package health

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bpineau/katafygio/config"
)

type healthHandler struct {
	config *config.KdnConfig
	donech chan struct{}
	srv    *http.Server
}

func New(config *config.KdnConfig) *healthHandler {
	return &healthHandler{
		config: config,
		donech: make(chan struct{}),
		srv:    nil,
	}
}

func (h *healthHandler) healthCheckReply(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "ok\n"); err != nil {
		h.config.Logger.Warningf("Failed to reply to http healtcheck from %s: %s\n", r.RemoteAddr, err)
	}
}

// Start exposes an http healthcheck handler
func (h *healthHandler) Start() (*healthHandler, error) {
	if h.config.HealthPort == 0 {
		return h, nil
	}

	h.srv = &http.Server{Addr: fmt.Sprintf(":%d", h.config.HealthPort)}

	http.HandleFunc("/health", h.healthCheckReply)

	go func() {
		defer close(h.donech)
		_ = h.srv.ListenAndServe()
	}()

	return h, nil
}

func (h *healthHandler) Stop() {
	h.config.Logger.Info("Stopping http healtcheck handler")
	if h.srv == nil {
		return
	}

	err := h.srv.Shutdown(context.TODO())
	if err != nil {
		h.config.Logger.Warningf("failed to stop http healtcheck handler: %v", err)
	}

	<-h.donech
}
