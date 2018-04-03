package controller

import (
	"fmt"
	"time"

	"github.com/bpineau/katafygio/config"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/ghodss/yaml"
)

const maxProcessRetry = 6

// Action represents the kind of object change we're notifying
type Action int

const (
	// Delete is the object deletion Action
	Delete Action = iota

	// Upsert is the update or create Action
	Upsert
)

// Event conveys an object delete/upsert notification
type Event struct {
	Action Action
	Key    string
	Kind   string
	Obj    string
}

// Controller is a generic kubernetes controller
type Controller struct {
	evchan   chan Event
	name     string
	config   *config.KdnConfig
	queue    workqueue.RateLimitingInterface
	informer cache.SharedIndexInformer
}

// NewController return an untyped, generic Kubernetes controller
func NewController(lw cache.ListerWatcher, evchan chan Event, name string, config *config.KdnConfig) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	informer := cache.NewSharedIndexInformer(
		lw,
		&unstructured.Unstructured{},
		config.ResyncIntv,
		cache.Indexers{},
	)

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

	return &Controller{evchan, name, config, queue, informer}
}

// Run starts the controller in the foreground
func (c *Controller) Run(stopCh <-chan struct{}) {
	c.config.Logger.Infof("Starting %s controller", c.name)
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
	// XXX needs a sync.wg to wait for that
}

func (c *Controller) runWorker() {
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

	err := c.processItem(key.(string))

	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxProcessRetry {
		c.config.Logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.config.Logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
	}

	return true
}

func (c *Controller) processItem(key string) error {
	rawobj, exists, err := c.informer.GetIndexer().GetByKey(key)

	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		// deleted object
		c.enqueue(Event{Action: Delete, Key: key, Kind: c.name, Obj: ""})
		return nil
	}

	obj := rawobj.(*unstructured.Unstructured).DeepCopy()

	// clean non exportable fields
	uc := obj.UnstructuredContent()
	md := uc["metadata"].(map[string]interface{})
	delete(uc, "status")
	delete(md, "selfLink")
	delete(md, "uid")
	delete(md, "resourceVersion")
	delete(md, "generation")

	c.config.Logger.Debugf("Found %s/%s [%s]", obj.GetAPIVersion(), obj.GetKind(), key)

	yml, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", key, err)
	}

	c.enqueue(Event{Action: Upsert, Key: key, Kind: c.name, Obj: string(yml)})
	return nil
}

func (c *Controller) enqueue(ev Event) {
	c.evchan <- ev
}
