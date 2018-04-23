package config

import (
	"fmt"
	"time"

	"github.com/bpineau/katafygio/pkg/client"

	"github.com/sirupsen/logrus"
)

// KfConfig holds the configuration options passed at launch time (and the rest client)
type KfConfig struct {
	// When DryRun is true, we don't write to disk and we don't commit/push
	DryRun bool

	// When DumpMode is true, we just dump everything once and exit
	DumpMode bool

	// Logger should be used to send all logs
	Logger *logrus.Logger

	// Client represents a connection to a Kubernetes cluster
	Client client.Interface

	// GitURL is the address of a remote git repository
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
	c.Client, err = client.New(apiserver, kubeconfig)
	if err != nil {
		return fmt.Errorf("Failed init Kubernetes client: %+v", err)
	}
	return nil
}
