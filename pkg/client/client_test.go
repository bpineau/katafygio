package client

import (
	"fmt"
	"os"
	"testing"
)

const nonExistentPath = "\\/hopefully/non/existent/path"

func TestClientSet(t *testing.T) {
	here, _ := os.Getwd()
	_ = os.Setenv("HOME", here+"/../../assets")
	cs, err := NewClientSet("", "")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%T", cs) != "*kubernetes.Clientset" {
		t.Errorf("NewClientSet() didn't return a *kubernetes.Clientset: %T", cs)
	}

	cs, _ = NewClientSet("http://127.0.0.1", "/dev/null")
	if fmt.Sprintf("%T", cs) != "*kubernetes.Clientset" {
		t.Errorf("NewClientSet(server) didn't return a *kubernetes.Clientset: %T", cs)
	}

	_, err = NewClientSet("http://127.0.0.1", nonExistentPath)
	if err == nil {
		t.Fatal("NewClientSet() should fail on non existent kubeconfig path")
	}

	_ = os.Unsetenv("KUBERNETES_SERVICE_HOST")
	_ = os.Setenv("HOME", nonExistentPath)
	_, err = NewClientSet("", "")
	if err == nil {
		t.Fatal("NewClientSet() should fail to load InClusterConfig without kube address env")
	}
}
