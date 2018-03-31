package config

import (
	"fmt"
	"time"

	"github.com/bpineau/katafygio/pkg/clientset"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KdnConfig is the configuration struct, passed to controllers's Init()
type KdnConfig struct {
	// When DryRun is true, we display but don't really send notifications
	DryRun bool

	// Logger should be used to send all logs
	Logger *logrus.Logger

	// ClientSet represents a connection to a Kubernetes cluster
	ClientSet kubernetes.Interface

	// GitUrl is the address of the git repository
	GitUrl string

	// LocalDir is the local path where we'll serialize cluster objets
	LocalDir string

	// Filter holds a facultative Kubernetes selector
	Filter string

	// Exclude holds a list of resources types we won't dump
	Exclude []string

	// ExcludeObj holds a list of objects we won't dump
	ExcludeObj []string

	// HealthPort is the facultative healthcheck port
	HealthPort int

	// ResyncIntv define the duration between full resync. Set to 0 to disable resyncs.
	ResyncIntv time.Duration
}

// Init initialize the configuration's ClientSet
func (c *KdnConfig) Init(apiserver string, kubeconfig string) error {
	var err error

	if c.ClientSet == nil {
		c.ClientSet, err = clientset.NewClientSet(apiserver, kubeconfig)
		if err != nil {
			return fmt.Errorf("Failed init Kubernetes clientset: %+v", err)
		}
	}

	// better fail early, if we can't talk to the cluster's api
	_, err = c.ClientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Failed to query Kubernetes api-server: %+v", err)
	}

	c.Logger.Info("Kubernetes clientset initialized")
	return nil
}
