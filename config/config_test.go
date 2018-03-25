package config

import (
	"fmt"
	"os"
	"testing"
)

const nonExistentPath = "\\/hopefully/non/existent/path"

func TestConfig(t *testing.T) {
	conf := FakeConfig()
	if conf == nil {
		t.Error("Failed to initialize a fake config object")
	}

	// test config's provided FakeClientSet we'll use throughout this file
	cs := FakeClientSet()
	if fmt.Sprintf("%T", cs) != "*fake.Clientset" {
		t.Errorf("FakeClientSet() failed")
	}

	// test with the fake clientset (should panic on error)
	err := conf.Init("", "")
	if err != nil {
		t.Errorf("Failed to initialize conf: %+v", err)
	}
	if fmt.Sprintf("%T", conf.ClientSet) != "*fake.Clientset" {
		t.Errorf("conf.Init() shouldn't overwrite an existing ClientSet")
	}

	// test with a real clientset
	conf.ClientSet = nil
	here, _ := os.Getwd()
	_ = os.Setenv("HOME", here+"/../..")
	_ = conf.Init("http://127.0.0.1", "/dev/null")
	if fmt.Sprintf("%T", conf.ClientSet) != "*kubernetes.Clientset" {
		t.Errorf("Should have a real *kubernetes.Clientset")
	}

	// ensure we raise an error if the provided config file is unreachable
	conf.ClientSet = nil
	err = conf.Init("http://127.0.0.1", nonExistentPath)
	if err == nil {
		t.Fatal("conf.Init() should fail on non existent kubeconfig path")
	}
}
