package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/log"
)

func TestNoopHealth(t *testing.T) {

	conf := &config.KfConfig{
		Logger:     log.New("error", "", "test"),
		HealthPort: 0,
	}

	// shouldn't panic with 0 as port
	hc := New(conf)
	_ = hc.Start()
	hc.Stop()

	conf.HealthPort = -42
	hc = New(conf)
	_ = hc.Start()
	hc.Stop()
	hook := hc.config.Logger.Hooks[logrus.InfoLevel][0].(*test.Hook)
	if len(hook.Entries) != 1 {
		t.Error("Failed to log an issue with a bogus port")
	}
}

func TestHealthCheck(t *testing.T) {
	conf := &config.KfConfig{
		Logger:     log.New("info", "", "test"),
		HealthPort: 0,
	}

	hc := New(conf)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(hc.healthCheckReply)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("healthCheckReply handler didn't return an HTTP 200 status code")
	}
}
