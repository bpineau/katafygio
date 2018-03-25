package log

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var (
	levels = [...]string{
		"debug",
		"info",
		"warning",
		"error",
		"fatal",
		"panic",
	}
)

func TestLog(t *testing.T) {
	logger := New("warning", "", "test")

	logger.Info("Changed: foo")
	logger.Warn("Changed: bar")
	logger.Error("Deleted: foo")

	hook := logger.Hooks[logrus.InfoLevel][0].(*test.Hook)
	if len(hook.Entries) != 2 {
		t.Errorf("Not the correct count of log entries")
	}

	logger.Warn("Changed: baz")
	if hook.LastEntry().Message != "Changed: baz" {
		t.Errorf("Unexpected log entry: %s", hook.LastEntry().Message)
	}

	logger = New("", "", "test")
	if logger.Level != logrus.InfoLevel {
		t.Error("The default loglevel should be info")
	}

	logger = New("", "", "")
	if logger.Out != os.Stderr {
		t.Error("The default output should be stderr")
	}

	logger = New("info", "127.0.0.1:514", "syslog")
	if fmt.Sprintf("%T", logger) != "*logrus.Logger" {
		t.Error("Failed to instantiate a syslog logger")
	}

	logger = New("info", "", "stdout")
	if fmt.Sprintf("%T", logger) != "*logrus.Logger" {
		t.Error("Failed to instantiate a stdout logger")
	}

	logger = New("info", "", "stderr")
	if fmt.Sprintf("%T", logger) != "*logrus.Logger" {
		t.Error("Failed to instantiate a stderr logger")
	}

	for _, level := range levels {
		lg := New(level, "", "test")
		if fmt.Sprintf("%T", lg) != "*logrus.Logger" {
			t.Errorf("Failed to instantiate at %s level", level)
		}
	}
}

func TestSyslogMissingArg(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("syslog logger should panic without a server")
		}
	}()

	_ = New("info", "", "syslog")
}

func TestSyslogWrongArg(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("syslog logger should panic on wrong server address")
		}
	}()

	_ = New("info", "wrong server", "syslog")
}
