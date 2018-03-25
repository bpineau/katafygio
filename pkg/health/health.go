// Package health serves healthchecks over HTTP at /health endpoint.
package health

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bpineau/katafygio/config"
)

type healthHandler struct {
	conf *config.KdnConfig
}

func (h *healthHandler) healthCheckReply(w http.ResponseWriter, r *http.Request) {
	if _, err := io.WriteString(w, "ok\n"); err != nil {
		h.conf.Logger.Warningf("Failed to reply to http healtcheck from %s: %s\n", r.RemoteAddr, err)
	}
}

// HeartBeatService exposes an http healthcheck handler
func HeartBeatService(c *config.KdnConfig) error {
	if c.HealthPort == 0 {
		return nil
	}
	hh := healthHandler{conf: c}
	http.HandleFunc("/health", hh.healthCheckReply)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.HealthPort), nil)
}
