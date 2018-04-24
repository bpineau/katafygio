// Package observer polls the Kubernetes api-server to discover all supported
// API groups/object kinds, and launch a new controller for each of them.
// Due to CRD/TPR, new API groups / object kinds may appear at any time,
// that's why we keep polling the API server.
package observer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/event"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

const discoveryInterval = 60 * time.Second

// ControllerFactory make controllers generation interchangeable
type ControllerFactory interface {
	NewController(client cache.ListerWatcher, notifier event.Notifier,
		name string, config *config.KfConfig) controller.Interface
}

type controllerCollection map[string]controller.Interface

// Observer watch api-server and manage kubernetes controllers lifecyles
type Observer struct {
	sync.RWMutex // protect ctrls
	config       *config.KfConfig
	stopCh       chan struct{}
	doneCh       chan struct{}
	notifier     event.Notifier
	discovery    discovery.DiscoveryInterface
	cpool        dynamic.ClientPool
	ctrls        controllerCollection
	factory      ControllerFactory
}

type gvk struct {
	groupVersion schema.GroupVersion
	apiResource  metav1.APIResource
}

type resources map[string]*gvk

// New returns a new observer, that will watch API resources and create controllers
func New(config *config.KfConfig, notif event.Notifier, factory ControllerFactory) *Observer {
	return &Observer{
		config:    config,
		notifier:  notif,
		discovery: discovery.NewDiscoveryClientForConfigOrDie(config.Client.GetRestConfig()),
		cpool:     dynamic.NewDynamicClientPool(config.Client.GetRestConfig()),
		ctrls:     make(controllerCollection),
		factory:   factory,
	}
}

// Start starts the observer in a detached goroutine
func (c *Observer) Start() *Observer {
	c.config.Logger.Info("Starting all kubernetes controllers")

	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(discoveryInterval)
		defer ticker.Stop()
		defer close(c.doneCh)

		for {
			err := c.refresh()
			if err != nil {
				c.config.Logger.Errorf("Refresh failed: %v", err)
			}

			select {
			case <-c.stopCh:
				return
			case <-ticker.C:
			}
		}
	}()

	return c
}

// Stop halts the observer
func (c *Observer) Stop() {
	c.config.Logger.Info("Stopping all kubernetes controllers")

	c.stopCh <- struct{}{}

	c.RLock()
	for _, ct := range c.ctrls {
		ct.Stop()
	}
	c.RUnlock()

	<-c.doneCh
}

func (c *Observer) refresh() error {
	c.Lock()
	defer c.Unlock()

	groups, err := c.discovery.ServerResources()
	if err != nil {
		return fmt.Errorf("failed to collect server resources: %v", err)
	}

	for name, res := range c.expandAndFilterAPIResources(groups) {
		if _, ok := c.ctrls[name]; ok {
			continue
		}

		kind := res.apiResource.Kind
		gk := res.groupVersion.WithKind(kind)
		cname := strings.ToLower(kind)

		cl, err := c.cpool.ClientForGroupVersionKind(gk)
		if err != nil {
			return fmt.Errorf("failed to get a client for %s", name)
		}

		client := cl.Resource(res.apiResource.DeepCopy(), metav1.NamespaceAll)

		c.ctrls[name] = c.factory.NewController(client, c.notifier, cname, c.config)
		go c.ctrls[name].Start()
	}

	return nil
}

// The api-server may expose a resource under several API groups, for backward
// compatibility. We'll want to ignore lower priorities "cohabitations":
// cf. kubernetes/cmd/kube-apiserver/app/server.go
var preferredVersions = map[string]string{
	"apps:deployment":                   "extensions:deployment",
	"apps:daemonset":                    "extensions:daemonset",
	"apps:replicaset":                   "extensions:replicaset",
	"events:events":                     ":events",
	"extensions:podsecuritypolicies":    "policy:podsecuritypolicies",
	"networking.k8s.io:networkpolicies": "extensions:networkpolicies",
}

func (c *Observer) expandAndFilterAPIResources(groups []*metav1.APIResourceList) resources {
	resources := make(map[string]*gvk)

	for _, group := range groups {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			c.config.Logger.Errorf("unparsable group version: %v", err)
			continue
		}

		for _, ar := range group.APIResources {
			// remove subresources (like job/status)
			if strings.ContainsRune(ar.Name, '/') {
				continue
			}

			// remove user filtered objet kinds
			if isExcluded(c.config.ExcludeKind, ar.Kind) {
				continue
			}

			// only consider resources that are getable, listable an watchable
			if !isSubList(ar.Verbs, []string{"list", "get", "watch"}) {
				continue
			}

			resources[strings.ToLower(gv.Group+":"+ar.Kind)] = &gvk{
				groupVersion: gv,
				apiResource:  ar,
			}
		}
	}

	for preferred, obsolete := range preferredVersions {
		if _, ok := resources[preferred]; ok {
			delete(resources, obsolete)
		}
	}

	return resources
}

func isExcluded(excluded []string, name string) bool {
	lname := strings.ToLower(name)
	for _, ctl := range excluded {
		if strings.Compare(lname, strings.ToLower(ctl)) == 0 {
			return true
		}
	}

	return false
}

func isSubList(containing []string, contained []string) bool {
containing:
	for _, in := range contained {
		for _, out := range containing {
			if strings.Compare(in, out) == 0 {
				continue containing
			}
		}
		return false
	}
	return true
}
