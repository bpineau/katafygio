// Package client  initialize a Kubernete's client-go rest.Config
package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Ensure we have GCP auth method linked in
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// BuildConfig create a *rest.Config
func BuildConfig(apiserver string, kubeconfig string) (*rest.Config, error) {
	// if we're not provided a kubeconfig path, try to find one in user's home
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			if _, err := os.Stat(filepath.Join(home, ".kube", "config")); err == nil {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}
	}

	// if we're provided an api-server url, or found a kubeconfig file, use that
	if apiserver != "" || kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	}

	// else assume we're running in cluster
	return rest.InClusterConfig()
}

// NewClientSet create a clientset (a client connection to a Kubernetes cluster).
// It will connect using the optional apiserver or kubeconfig options, or will
// default to the automatic, in cluster settings.
func NewClientSet(apiserver string, kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := BuildConfig(apiserver, kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
