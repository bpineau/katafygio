package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/bpineau/katafygio/pkg/log"
)

var logs = log.New("error", "", "test")

func TestNoopHealth(t *testing.T) {

	// shouldn't panic with 0 as port
	hc := New(logs, 0)
	_ = hc.Start()
	hc.Stop()

	hc = New(logs, -42)
	_ = hc.Start()
	hc.Stop()
	hook := logs.Hooks[logrus.InfoLevel][0].(*test.Hook)
	if len(hook.Entries) != 1 {
		t.Error("Failed to log an issue with a bogus port")
	}
}

func TestHealthCheck(t *testing.T) {
	hc := New(logs, 0)

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
