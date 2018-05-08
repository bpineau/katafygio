package client

import (
	"fmt"
	"os"
	"testing"
)

const nonExistentPath = "\\/non / existent / $path$"

func TestClientSet(t *testing.T) {
	here, _ := os.Getwd()
	_ = os.Setenv("HOME", here+"/../../assets")
	cs, err := New("", "")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%T", cs.GetRestConfig()) != "*rest.Config" {
		t.Errorf("GetRestConfig() didn't return a *rest.Config: %T", cs)
	}

	cs, _ = New("http://127.0.0.1", "/dev/null")
	if fmt.Sprintf("%T", cs.GetRestConfig()) != "*rest.Config" {
		t.Errorf("New(server) didn't return a *rest.Config: %T", cs)
	}

	_, err = New("http://127.0.0.1", nonExistentPath)
	if err == nil {
		t.Fatal("New() should fail on non existent kubeconfig path")
	}

	_ = os.Unsetenv("KUBERNETES_SERVICE_HOST")
	_ = os.Setenv("HOME", nonExistentPath)
	_ = os.Setenv("KUBECONFIG", nonExistentPath)
	_, err = New("", "")
	if err == nil {
		t.Fatal("New() should fail to load InClusterConfig without kube address env")
	}
}
