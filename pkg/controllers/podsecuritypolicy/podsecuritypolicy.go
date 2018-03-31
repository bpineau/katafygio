package podsecuritypolicy

import (
	"fmt"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controllers"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/ghodss/yaml"
)

// Controller monitors Kubernetes' objects in the cluster
type Controller struct {
	controllers.CommonController
}

func init() {
	controllers.Register("podsecuritypolicy", New)
}

// New initialize controller
func New(conf *config.KdnConfig, ch chan<- controllers.Event) controllers.Controller {
	c := Controller{}
	c.CommonController = controllers.CommonController{
		Conf: conf,
		Name: "podsecuritypolicy",
		Send: ch,
	}

	client := c.Conf.ClientSet
	c.ObjType = &v1beta1.PodSecurityPolicy{}
	c.MarshalF = c.Marshal // that's our own Marshal() implementation
	selector := meta_v1.ListOptions{LabelSelector: conf.Filter}
	c.ListWatch = &cache.ListWatch{
		ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
			return client.ExtensionsV1beta1().PodSecurityPolicies().List(selector)
		},
		WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
			return client.ExtensionsV1beta1().PodSecurityPolicies().Watch(selector)
		},
	}

	return &c
}

// Marshal filter irrelevant fields from the object, and export it as yaml string
func (c *Controller) Marshal(obj interface{}) (string, error) {
	f := obj.(*v1beta1.PodSecurityPolicy).DeepCopy()

	// some attributes, added by the cluster, shouldn't be exported
	//f.Status.Reset()
	f.ResourceVersion = ""
	f.SelfLink = ""
	f.UID = ""

	y, err := yaml.Marshal(f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("apiVersion: %s\nkind: %s\n%s",
		c.Conf.ClientSet.ExtensionsV1beta1().RESTClient().APIVersion().String(),
		"PodSecurityPolicy",
		string(y)), nil
}
