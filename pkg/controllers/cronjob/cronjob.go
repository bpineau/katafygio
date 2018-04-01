package cronjob

import (
	"fmt"
	"strings"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controllers"

	batchv1 "k8s.io/api/batch/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/ghodss/yaml"
)

const objectKind = "CronJob"

// Controller monitors Kubernetes' objects in the cluster
type Controller struct {
	controllers.CommonController
}

func init() {
	controllers.Register(strings.ToLower(objectKind), New)
}

// New initialize controller
func New(conf *config.KdnConfig, ch chan<- controllers.Event) controllers.Controller {
	c := Controller{}
	c.CommonController = controllers.CommonController{
		Conf: conf,
		Name: strings.ToLower(objectKind),
		Send: ch,
	}

	client := c.Conf.ClientSet
	c.ObjType = &batchv1.CronJob{}
	c.MarshalF = c.Marshal
	selector := meta.ListOptions{LabelSelector: conf.Filter}
	c.ListWatch = &cache.ListWatch{
		ListFunc: func(options meta.ListOptions) (runtime.Object, error) {
			return client.BatchV1beta1().CronJobs(meta.NamespaceAll).List(selector)
		},
		WatchFunc: func(options meta.ListOptions) (watch.Interface, error) {
			return client.BatchV1beta1().CronJobs(meta.NamespaceAll).Watch(selector)
		},
	}

	return &c
}

// Marshal filter irrelevant fields from the object, and export it as yaml string
func (c *Controller) Marshal(obj interface{}) (string, error) {
	f := obj.(*batchv1.CronJob).DeepCopy()

	f.Status.Reset()
	f.ResourceVersion = ""
	f.SelfLink = ""
	f.UID = ""

	y, err := yaml.Marshal(f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("apiVersion: %s\nkind: %s\n%s",
		batchv1.SchemeGroupVersion.String(), objectKind, string(y)), nil
}
