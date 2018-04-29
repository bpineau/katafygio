package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"
	"k8s.io/client-go/rest"
)

type mockClient struct{}

func (m *mockClient) GetRestConfig() *rest.Config {
	return &rest.Config{}
}

func TestRootCmd(t *testing.T) {
	restcfg = new(mockClient)
	appFs = afero.NewMemMapFs()
	RootCmd.SetOutput(new(bytes.Buffer))
	RootCmd.SetArgs([]string{
		"--config",
		"/dev/null",
		"--kube-config",
		"/dev/null",
		"--dry-run",
		"--dump-only",
		"--api-server",
		"http://192.0.2.1", // RFC 5737 reserved/unroutable
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

	if err := Execute(); err != nil {
		t.Errorf("version subcommand shouldn't fail: %+v", err)
	}
}

func TestVersionCmd(t *testing.T) {
	RootCmd.SetOutput(new(bytes.Buffer))
	RootCmd.SetArgs([]string{"version"})
	if err := RootCmd.Execute(); err != nil {
		t.Errorf("version subcommand shouldn't fail: %+v", err)
	}
}
