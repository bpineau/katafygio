// Package controllers is responsible for watching resources.
// Each controller (implementing the Controller interface)
// watchs for a specific Kubernetes object (ie. deployments).
package controllers

import (
	"fmt"
	"sync"
	"time"

	"github.com/bpineau/katafygio/config"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	maxProcessRetry = 6
)

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

// Controller are started in a persistent goroutine at program launch,
// and are responsible for watching resources.
type Controller interface {
	Start(wg *sync.WaitGroup)
	Stop()
	//Init(c *config.KdnConfig, ch chan<- Event) Controller
	Marshal(obj interface{}) (string, error)
}

// MashalFunc serialize a Kube object as an exportable yaml string.
type MashalFunc func(obj interface{}) (string, error)

// CommonController implements the core reusable and generic primitives
// of a controller, and can be embedded by real controllers.
type CommonController struct {
	Conf      *config.KdnConfig
	Queue     workqueue.RateLimitingInterface
	Informer  cache.SharedIndexInformer
	Name      string
	ListWatch cache.ListerWatcher
	ObjType   runtime.Object
	StopCh    chan struct{}
	wg        *sync.WaitGroup
	initMu    sync.Mutex
	syncInit  bool
	MarshalF  MashalFunc
	Send      chan<- Event
}

// Concrete controllers implementations registration

// ControllerConstructor creates a new controller
type ControllerConstructor func(conf *config.KdnConfig, ch chan<- Event) Controller

// AllControllers is the registration bucket for all controllers
var AllControllers = make(map[string]ControllerConstructor)

// Register is call by controllers init()
func Register(name string, ctrl ControllerConstructor) {
	AllControllers[name] = ctrl
}

// Start initialize and launch a controller. The sync.WaitGroup
// argument is expected to be aknowledged (Done()) at controller
// termination, when Stop() is called.
func (c *CommonController) Start(wg *sync.WaitGroup) {
	c.Conf.Logger.Infof("Starting %s controller", c.Name)

	c.StopCh = make(chan struct{})

	c.wg = wg

	c.initMu.Lock()
	c.syncInit = true
	c.initMu.Unlock()

	c.startInformer()

	go c.run(c.StopCh)

	<-c.StopCh
}

// Stop ends a controller and notify the controller's WaitGroup
func (c *CommonController) Stop() {
	c.Conf.Logger.Infof("Stopping %s controller", c.Name)

	// don't stop while we're still starting
	c.initMu.Lock()
	for !c.syncInit {
		time.Sleep(time.Millisecond)
	}
	c.initMu.Unlock()

	close(c.StopCh)

	// give everything 0.2s max to stop gracefully
	time.Sleep(200 * time.Millisecond)

	c.wg.Done()
}

func (c *CommonController) startInformer() {
	c.Queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	c.Informer = cache.NewSharedIndexInformer(
		c.ListWatch,
		c.ObjType,
		c.Conf.ResyncIntv,
		cache.Indexers{},
	)

	c.Informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.Queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.Queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.Queue.Add(key)
			}
		},
	})
}

func (c *CommonController) run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.Queue.ShutDown()

	go c.Informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.Informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	c.Conf.Logger.Infof("%s controller synced and ready", c.Name)

	wait.Until(c.runWorker, time.Second, stopCh)
}

func (c *CommonController) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *CommonController) processNextItem() bool {
	key, quit := c.Queue.Get()
	if quit {
		return false
	}
	defer c.Queue.Done(key)

	err := c.processItem(key.(string))

	if err == nil {
		// No error, reset the ratelimit counters
		c.Queue.Forget(key)
	} else if c.Queue.NumRequeues(key) < maxProcessRetry {
		c.Conf.Logger.Errorf("Error processing %s (will retry): %v", key, err)
		c.Queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		c.Conf.Logger.Errorf("Error processing %s (giving up): %v", key, err)
		c.Queue.Forget(key)
	}

	return true
}

func (c *CommonController) processItem(key string) error {
	obj, exists, err := c.Informer.GetIndexer().GetByKey(key)

	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		c.enqueue(Event{Action: Delete, Key: key, Kind: c.Name, Obj: ""})
		return nil
	}

	yml, err := c.MarshalF(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %v", key, err)
	}

	c.enqueue(Event{Action: Upsert, Key: key, Kind: c.Name, Obj: yml})
	return nil
}

func (c *CommonController) enqueue(ev Event) {
	c.Send <- ev
}
