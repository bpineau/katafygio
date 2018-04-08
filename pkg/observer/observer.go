// Package observer polls the Kubernetes api-server to discover all supported
// API groups/object kinds, and launch a new controller for each of them.
// Due to CRD/TPR, new API groups / object kinds may appear at any time,
// that's why we keep polling the API server.
package observer

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/event"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

const discoveryInterval = 60 * time.Second

// Observer watch api-server and manage kubernetes controllers
type Observer struct {
	stop   chan struct{}
	done   chan struct{}
	notif  *event.Notifier
	disc   *discovery.DiscoveryClient
	cpool  dynamic.ClientPool
	ctrls  map[string]*controller.Controller
	config *config.KfConfig
}

type gvk struct {
	group   string
	version string
	kind    string
	gv      schema.GroupVersion
	ar      metav1.APIResource
}

type resources map[string]*gvk

// New returns a new observer, that will watch for api resource kinds
// and create new controllers for each one.
func New(config *config.KfConfig, notif *event.Notifier) *Observer {
	return &Observer{
		config: config,
		notif:  notif,
		disc:   discovery.NewDiscoveryClientForConfigOrDie(config.Client),
		cpool:  dynamic.NewDynamicClientPool(config.Client),
		ctrls:  make(map[string]*controller.Controller),
	}
}

// Start starts the observer in a detached goroutine
func (c *Observer) Start() *Observer {
	c.config.Logger.Info("Starting all kubernetes controllers")

	c.stop = make(chan struct{})
	c.done = make(chan struct{})

	go func() {
		ticker := time.NewTicker(discoveryInterval)
		defer ticker.Stop()
		defer close(c.done)

		for {
			err := c.refresh()
			if err != nil {
				c.config.Logger.Errorf("Failed to refresh: %v", err)
			}

			select {
			case <-c.stop:
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

	close(c.stop)

	for _, c := range c.ctrls {
		c.Stop()
	}

	<-c.done
}

func (c *Observer) refresh() error {
	groups, err := c.disc.ServerResources()
	if err != nil {
		return fmt.Errorf("failed to collect server resources: %v", err)
	}

	for name, res := range c.expandAndFilterAPIResources(groups) {
		if _, ok := c.ctrls[name]; ok {
			continue
		}

		cl, err := c.cpool.ClientForGroupVersionKind(res.gv.WithKind(res.kind))
		if err != nil {
			return fmt.Errorf("failed to get a cpool for %s", name)
		}

		client := cl.Resource(res.ar.DeepCopy(), metav1.NamespaceAll)

		selector := metav1.ListOptions{LabelSelector: c.config.Filter}
		lw := &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return client.List(selector)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return client.Watch(selector)
			},
		}

		c.ctrls[name] = controller.New(lw, c.notif, strings.ToLower(res.ar.Kind), c.config)
		go c.ctrls[name].Start()
	}

	return nil
}

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
			c.config.Logger.Errorf("api-server sent an unparsable group version: %v", err)
			continue
		}

		for _, ar := range group.APIResources {
			// remove subresources (like job/status or deployments/scale)
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

			resource := &gvk{group: gv.Group, version: gv.Version, kind: ar.Kind, gv: gv, ar: ar}
			resources[strings.ToLower(gv.Group+":"+ar.Kind)] = resource
		}
	}

	// remove lower priorities "cohabitations". cf. kubernetes/cmd/kube-apiserver/app/server.go
	// (the api-server may expose a resource under several api groups, for backward compat)
	for preferred, obsolete := range preferredVersions {
		if _, ok := resources[preferred]; ok {
			delete(resources, obsolete)
		}
	}

	return resources
}

func isExcluded(excluded []string, name string) bool {
	for _, ctl := range excluded {
		if strings.Compare(strings.ToLower(name), strings.ToLower(ctl)) == 0 {
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
