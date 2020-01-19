// Package client initialize and wrap a Kubernete's client-go rest.Config client
package client

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

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
func New(apiserver, context, kubeconfig string) (*RestClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{
			ClusterInfo:    clientcmdapi.Cluster{Server: apiserver},
			CurrentContext: context,
		},
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build a restconfig: %v", err)
	}

	return &RestClient{
		cfg: restConfig,
	}, nil
}

// GetRestConfig returns the current rest.Config
func (r *RestClient) GetRestConfig() *rest.Config {
	return r.cfg
}
