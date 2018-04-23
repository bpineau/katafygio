// Package client initialize and wrap a Kubernete's client-go rest.Config client
package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Ensure we have auth plugins (gcp, azure, openstack, ...) linked in
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// Interface abstracts access to a concrete Kubernetes rest.Client
type Interface interface {
	GetRestConfig() *rest.Config
}

// RestClient holds a Kubernetes rest client configuration
type RestClient struct {
	cfg *rest.Config
}

// New create a new RestClient
func New(apiserver, kubeconfig string) (*RestClient, error) {
	cfg, err := newRestConfig(apiserver, kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build a restconfig: %v", err)
	}

	return &RestClient{
		cfg: cfg,
	}, nil
}

// GetRestConfig returns the current rest.Config
func (r *RestClient) GetRestConfig() *rest.Config {
	return r.cfg
}

// newRestConfig create a *rest.Config, trying to mimic kubectl behavior:
// - Explicit user provided api-server (and/or kubeconfig path) have higher priorities
// - Else, use the config file path in KUBECONFIG environment variable (if any)
// - Else, use the config file in ~/.kube/config, if any
// - Else, consider we're running in cluster (in a pod), and use the pod's service
//   account and cluster's kubernetes.default service.
func newRestConfig(apiserver string, kubeconfig string) (*rest.Config, error) {
	// if not passed as an argument, kubeconfig can be provided as env var
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	// if we're not provided an explicit kubeconfig path (via env, or argument),
	// try to find one at the standard place (in user's home/.kube/config).
	homeCfg := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(homeCfg)
	if kubeconfig == "" && err == nil {
		kubeconfig = homeCfg
	}

	// if we were provided or found a kubeconfig,
	// or if we were provided an api-server url, use that
	if apiserver != "" || kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	}

	// else assume we're running in a pod, in cluster
	return rest.InClusterConfig()
}
