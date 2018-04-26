package log

import (
	"fmt"
	"os"
	"runtime"
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
	logger, err := New("warning", "", "test")
	if err != nil {
		t.Errorf("Creating a new test logger shouldn't fail: %v", err)
	}

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

	for _, level := range levels {
		lg, err2 := New(level, "", "test")
		if err2 != nil || fmt.Sprintf("%T", lg) != "*logrus.Logger" {
			t.Errorf("Failed to instantiate at %s level: %v", level, err2)
		}
	}

	logger, err = New("", "", "test")
	if err != nil || logger.Level != logrus.InfoLevel {
		t.Errorf("The default loglevel should be info %v", err)
	}

	logger, err = New("", "", "")
	if err != nil || logger.Out != os.Stderr {
		t.Errorf("The default output should be stderr %v", err)
	}

	if runtime.GOOS != "windows" {
		logger, err = New("info", "127.0.0.1:514", "syslog")
		if err != nil || fmt.Sprintf("%T", logger) != "*logrus.Logger" {
			t.Errorf("Failed to instantiate a syslog logger %v", err)
		}
	}

	logger, err = New("info", "", "stdout")
	if err != nil || fmt.Sprintf("%T", logger) != "*logrus.Logger" {
		t.Errorf("Failed to instantiate a stdout logger %v", err)
	}

	logger, err = New("info", "", "stderr")
	if err != nil || fmt.Sprintf("%T", logger) != "*logrus.Logger" {
		t.Errorf("Failed to instantiate a stderr logger %v", err)
	}
}

func TestSyslogMissingArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("syslog is not supported on Windows")
	}

	_, err := New("info", "", "syslog")
	if err == nil {
		t.Errorf("syslog logger should fail without a server")
	}
}

func TestSyslogWrongArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("syslog is not supported on Windows")
	}

	_, err := New("info", "wrong server", "syslog")
	if err == nil {
		t.Errorf("syslog logger should fail with a broken server address")
	}
}
