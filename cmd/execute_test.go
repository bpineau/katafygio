package cmd

import (
	"bytes"
	"syscall"
	"testing"
	"time"
)

// most of cli binding code is executed through the magical init() mecanism
func TestRootCmd(t *testing.T) {
	FakeCS = true
	RootCmd.SetOutput(new(bytes.Buffer))
	RootCmd.SetArgs([]string{
		"--config",
		"/dev/null",
		"--dry-run",
		"--api-server",
		"http://127.0.0.1",
		"--log-level",
		"warning",
		"--log-output",
		"test",
		"--healthcheck-port",
		"0",
		"--filter",
		"foo=bar,spam=egg",
		"--resync-interval",
		"1",
	})

	ch := make(chan error, 1)

	go func() {
		ch <- Execute()
	}()

	select {
	case err := <-ch:
		if err != nil {
			t.Errorf("Failed to execute the main command: %+v", err)
		}
	case <-time.After(time.Second):
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	case <-time.After(10 * time.Second):
		t.Error("Timeout waiting for the execute command to exit after SIGTERM")
	}

	FakeCS = false

	RootCmd.SetArgs([]string{
		"--dry-run",
		"--api-server",
		"http://127.0.0.1",
		"--config",
		"\\/non/existent/path",
	})
	if err := Execute(); err == nil {
		t.Error("Execute() should fail with unreachable config file path")
	}

	RootCmd.SetArgs([]string{
		"--dry-run",
		"--api-server",
		"http://127.0.0.1",
		"--config",
	})
	if err := Execute(); err == nil {
		t.Error("Execute() should fail with missing flags arguments")
	}
}

func TestVersion(t *testing.T) {
	RootCmd.SetOutput(new(bytes.Buffer))
	RootCmd.SetArgs([]string{"version"})
	if err := RootCmd.Execute(); err != nil {
		t.Errorf("version subcommand shouldn't fail: %+v", err)
	}
}
