// Package observer polls the Kubernetes api-server to discover all supported
// API groups/object kinds, and launch a new controller for each of them.
// Due to CRD/TPR, new API groups / object kinds may appear at any time,
// that's why we keep polling the API server.
package observer

import (
	"strings"
	"sync"
	"time"

	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/event"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const discoveryInterval = 60 * time.Second

// ControllerFactory make controllers generation interchangeable
type ControllerFactory interface {
	NewController(client cache.ListerWatcher, notifier event.Notifier, name string) controller.Interface
}

type controllerCollection map[string]controller.Interface

type restclient interface {
	GetRestConfig() *rest.Config
}

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Observer watch api-server and manage kubernetes controllers lifecyles
type Observer struct {
	sync.RWMutex // protect ctrls
	stopCh       chan struct{}
	doneCh       chan struct{}
	notifier     event.Notifier
	discovery    discovery.DiscoveryInterface
	cpool        dynamic.Interface
	ctrls        controllerCollection
	factory      ControllerFactory
	logger       logger
	excludedkind []string
	namespace    string
}

type gvk struct {
	groupVersion schema.GroupVersion
	apiResource  metav1.APIResource
}

type resources map[string]*gvk

// New returns a new observer, that will watch API resources and create controllers
func New(log logger, client restclient, notif event.Notifier, factory ControllerFactory, excluded []string, namespace string) *Observer {
	return &Observer{
		notifier:     notif,
		discovery:    discovery.NewDiscoveryClientForConfigOrDie(client.GetRestConfig()),
		cpool:        dynamic.NewForConfigOrDie(client.GetRestConfig()),
		ctrls:        make(controllerCollection),
		factory:      factory,
		logger:       log,
		excludedkind: excluded,
		namespace:    namespace,
	}
}

// Start starts the observer in a detached goroutine
func (c *Observer) Start() *Observer {
	c.logger.Infof("Starting all kubernetes controllers")

	c.stopCh = make(chan struct{})
	c.doneCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(discoveryInterval)
		defer ticker.Stop()
		defer close(c.doneCh)

		for {
			err := c.refresh()
			if err != nil {
				c.logger.Errorf("Refresh failed: %v", err)
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
	c.logger.Infof("Stopping all kubernetes controllers")

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

	_, resources, err := c.discovery.ServerGroupsAndResources()
	if err != nil {
		c.logger.Errorf("failed to collect some server resources: %v", err)
	}

	for name, res := range c.expandAndFilterAPIResources(resources) {
		if _, ok := c.ctrls[name]; ok {
			continue
		}

		resource := schema.GroupVersionResource{
			Group:    res.groupVersion.Group,
			Version:  res.groupVersion.Version,
			Resource: res.apiResource.Name,
		}

		cname := strings.ToLower(res.apiResource.Kind)
		namespace := metav1.NamespaceAll
		if c.namespace != "" {
			namespace = c.namespace
		}
		lw := &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return c.cpool.Resource(resource).Namespace(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return c.cpool.Resource(resource).Namespace(namespace).Watch(options)
			},
		}

		c.ctrls[name] = c.factory.NewController(lw, c.notifier, cname)
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
			c.logger.Errorf("unparsable group version: %v", err)
			continue
		}

		for _, ar := range group.APIResources {
			// remove subresources (like job/status)
			if strings.ContainsRune(ar.Name, '/') {
				continue
			}

			// remove user filtered objet kinds
			if isExcluded(c.excludedkind, ar) {
				continue
			}

			// ignore non namespaced resources, when we have a namespace filter
			if c.namespace != "" && !ar.Namespaced {
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

func isExcluded(excluded []string, ar metav1.APIResource) bool {
	lname := strings.ToLower(ar.Name)
	lkind := strings.ToLower(ar.Kind)
	singular := strings.ToLower(ar.SingularName)

	for _, ctl := range excluded {
		excl := strings.ToLower(ctl)

		if strings.Compare(lname, excl) == 0 {
			return true
		}

		if strings.Compare(lkind, excl) == 0 {
			return true
		}

		if strings.Compare(singular, excl) == 0 {
			return true
		}

		for _, alt := range ar.ShortNames {
			if strings.Compare(strings.ToLower(alt), excl) == 0 {
				return true
			}
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
