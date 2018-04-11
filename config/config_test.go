package config

import (
	"os"
	"testing"

	"github.com/bpineau/katafygio/pkg/log"
)

const nonExistentPath = "\\/hopefully/non/existent/path"

func TestConfig(t *testing.T) {
	conf := &KfConfig{
		DryRun: true,
		Logger: log.New("info", "", "test"),
	}

	err := conf.Init("http://127.0.0.1", nonExistentPath)
	if err == nil {
		t.Error("conf.Init() should fail on non existent kubeconfig path")
	}

	here, _ := os.Getwd()
	_ = os.Setenv("HOME", here+"/../assets")
	err = conf.Init("", "")
	if err != nil {
		t.Error("conf.Init() with no arguments should find a .kube/config in $HOME")
	}

}
