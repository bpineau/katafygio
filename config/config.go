package config

import (
	"fmt"
	"time"

	"github.com/bpineau/katafygio/pkg/client"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

// KfConfig is the configuration struct, passed to controllers's Init()
type KfConfig struct {
	// When DryRun is true, we display but don't really send notifications
	DryRun bool

	// Logger should be used to send all logs
	Logger *logrus.Logger

	// Client represents a connection to a Kubernetes cluster
	Client *rest.Config

	// GitURL is the address of the git repository
	GitURL string

	// LocalDir is the local path where we'll serialize cluster objets
	LocalDir string

	// Filter holds a facultative Kubernetes selector
	Filter string

	// ExcludeKind holds a list of resources types we won't dump
	ExcludeKind []string

	// ExcludeObject holds a list of objects we won't dump
	ExcludeObject []string

	// HealthPort is the facultative healthcheck port
	HealthPort int

	// ResyncIntv define the duration between full resync. Set to 0 to disable resyncs.
	ResyncIntv time.Duration
}

// Init initialize the config
func (c *KfConfig) Init(apiserver string, kubeconfig string) (err error) {
	c.Client, err = client.NewRestConfig(apiserver, kubeconfig)
	if err != nil {
		return fmt.Errorf("Failed init Kubernetes clientset: %+v", err)
	}
	return nil
}
