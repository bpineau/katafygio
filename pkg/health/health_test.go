package health

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/log"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	conf := new(config.KdnConfig)
	hh := healthHandler{conf: conf}
	handler := http.HandlerFunc(hh.healthCheckReply)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("healthCheckReply handler didn't return an HTTP 200 status code")
	}

	if rr.Body.String() != "ok\n" {
		t.Errorf("healthCheckReply didn't return 'ok\n'")
	}

	if HeartBeatService(conf) != nil {
		t.Errorf("HeartBeatService should ignore unconfigured healthcheck")
	}

	conf.HealthPort = -42
	if HeartBeatService(conf) == nil {
		t.Errorf("HeartBeatService should fail with a wrong port")
	}

	hh.conf.Logger = log.New("warning", "", "test")
	hh.healthCheckReply(new(FailingResponseWriter), &http.Request{RemoteAddr: "127.0.0.1"})
	hook := hh.conf.Logger.Hooks[logrus.InfoLevel][0].(*test.Hook)
	if len(hook.Entries) != 1 {
		t.Error("Failed to log an issue while replying to healthcheck")
	}
}

type FailingResponseWriter struct{}

func (f *FailingResponseWriter) Write(b []byte) (int, error) {
	return 0, fmt.Errorf("Failed to write to socket")
}

func (f *FailingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (f *FailingResponseWriter) WriteHeader(i int) {
}
