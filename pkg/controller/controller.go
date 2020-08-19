// Package controller list and keep watching a specific Kubernetes resource kind
// (ie. "apps/v1 Deployment", "v1 Namespace", etc) and notifies a recorder whenever
// a change happens (an object changed, was created, or deleted). This is a generic
// implementation: the resource kind to watch is provided at runtime. We should
// start several such controllers to watch for distinct resources.
package controller

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bpineau/katafygio/pkg/event"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/ghodss/yaml"
)

var (
	maxProcessRetry = 6
	canaryKey       = "$katafygio canary$"
	unexported      = []string{"selfLink", "uid", "resourceVersion", "generation"}
)

// Interface describe a standard kubernetes controller
type Interface interface {
	Start()
	Stop()
}

type logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Factory generate controllers
type Factory struct {
	logger      logger
	filter      string
	resyncIntv  time.Duration
	excludedobj []string
	excludedns  []string
}

// Controller is a generic kubernetes controller
type Controller struct {
	name        string
	stopCh      chan struct{}
	doneCh      chan struct{}
	syncCh      chan struct{}
	notifier    event.Notifier
	queue       workqueue.RateLimitingInterface
	informer    cache.SharedIndexInformer
	logger      logger
	resyncIntv  time.Duration
	excludedobj []string
	excludedns  []*regexp.Regexp
}

// New return a kubernetes controller using the provided client
func New(client cache.ListerWatcher,
	notifier event.Notifier,
	log logger,
	name string,
	filter string,
	resync time.Duration,
	excludedobj []string,
	excludednamespace []string,
) *Controller {

	selector := metav1.ListOptions{LabelSelector: filter, ResourceVersion: "0"}
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.List(selector)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.Watch(selector)
		},
	}

	informer := cache.NewSharedIndexInformer(
		lw,
		&unstructured.Unstructured{},
		resync,
		cache.Indexers{},
	)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	})

	exclnsre := make([]*regexp.Regexp, 0)
	for _, ns := range excludednamespace {
		exclnsre = append(exclnsre, regexp.MustCompile(ns))
	}

	return &Controller{
		stopCh:      make(chan struct{}),
		doneCh:      make(chan struct{}),
		syncCh:      make(chan struct{}, 1),
		notifier:    notifier,
		name:        name,
		queue:       queue,
		informer:    informer,
		logger:      log,
		resyncIntv:  resync,
		excludedobj: excludedobj,
		excludedns:  exclnsre,
	}
}

// Start launchs the controller in the background
func (c *Controller) Start() {
	c.logger.Infof("Starting %s controller", c.name)
	defer utilruntime.HandleCrash()

	go c.informer.Run(c.stopCh)

	if !cache.WaitForCacheSync(c.stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for cache sync"))
		return
	}

	c.queue.Add(canaryKey)

	go wait.Until(c.runWorker, time.Second, c.stopCh)
}

// Stop halts the controller
func (c *Controller) Stop() {
	c.logger.Infof("Stopping %s controller", c.name)
	<-c.syncCh
	close(c.stopCh)
	c.queue.ShutDown()
	<-c.doneCh
}

func (c *Controller) runWorker() {
	defer close(c.doneCh)
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	if strings.Compare(key.(string), canaryKey) == 0 {
		c.logger.Infof("Initial sync completed for %s controller", c.name)
		c.syncCh <- struct{}{}
		c.queue.Forget(key)
		return true
	}

	err := c.processItem(key.(string))

	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxProcessRetry {
		c.logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	rawobj, exists, err := c.informer.GetIndexer().GetByKey(key)

	if err != nil {
		return fmt.Errorf("error fetching %s from store: %v", key, err)
	}

	for _, obj := range c.excludedobj {
		if strings.Compare(strings.ToLower(obj), strings.ToLower(c.name+":"+key)) == 0 {
			return nil
		}
	}

	if !exists {
		// deleted object
		c.enqueue(&event.Notification{Action: event.Delete, Key: key, Kind: c.name, Object: nil})
		return nil
	}

	obj := rawobj.(*unstructured.Unstructured).DeepCopy()

	// clear irrelevant attributes
	uc := obj.UnstructuredContent()
	delete(uc, "status")
	md := uc["metadata"].(map[string]interface{})
	for _, attr := range unexported {
		delete(md, attr)
	}

	if namespace, ok := md["namespace"].(string); ok {
		for _, nsre := range c.excludedns {
			if nsre.MatchString(namespace) {
				// Rely on the background sync to delete these excluded files if
				// we previously had aquired them
				return nil
			}
		}
	}

	yml, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", key, err)
	}

	c.enqueue(&event.Notification{Action: event.Upsert, Key: key, Kind: c.name, Object: yml})
	return nil
}

func (c *Controller) enqueue(notif *event.Notification) {
	c.notifier.Send(notif)
}

// NewFactory create a controller factory
func NewFactory(logger logger, filter string, resync int, excludedobj []string, excludedns []string) *Factory {
	return &Factory{
		logger:      logger,
		filter:      filter,
		resyncIntv:  time.Duration(resync) * time.Second,
		excludedobj: excludedobj,
		excludedns:  excludedns,
	}
}

// NewController create a controller.Controller
func (f *Factory) NewController(client cache.ListerWatcher, notifier event.Notifier, name string) Interface {
	return New(client, notifier, f.logger, name, f.filter, f.resyncIntv, f.excludedobj, f.excludedns)
}
