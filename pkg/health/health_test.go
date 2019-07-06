package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockLog struct {
	count int
}

func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Errorf(format string, args ...interface{}) { m.count++ }

var logs = new(mockLog)

func TestNoopHealth(t *testing.T) {

	// shouldn't panic with 0 as port
	hc := New(logs, 0)
	_ = hc.Start()
	hc.Stop()
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
