package observer

import (
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/bpineau/katafygio/pkg/controller"
	"github.com/bpineau/katafygio/pkg/event"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	fakerest "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/cache"
)

type mockNotifier struct{}

func (m *mockNotifier) Send(ev *event.Notification) {}
func (m *mockNotifier) ReadChan() <-chan event.Notification {
	return make(chan event.Notification)
}

type mockCtrl struct{}

func (m *mockCtrl) Start() {}
func (m *mockCtrl) Stop()  {}

type mockFactory struct {
	names []string
}

func (m *mockFactory) NewController(client cache.ListerWatcher, notifier event.Notifier, name string) controller.Interface {
	m.names = append(m.names, name)
	return &mockCtrl{}
}

type mockClient struct{}

func (m *mockClient) GetRestConfig() *rest.Config {
	return &rest.Config{}
}

type mockLog struct{}

func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Errorf(format string, args ...interface{}) {}

var stdVerbs = []string{"list", "get", "watch"}
var emptyExclude = make([]string, 0)

type resTest struct {
	title     string
	resources []*metav1.APIResourceList
	exclude   []string
	expect    []string
}

var resourcesTests = []resTest{

	{
		title:   "Normal resource parsing",
		exclude: emptyExclude,
		expect:  []string{"pod", "replicationcontroller", "replicaset"},
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: corev1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: stdVerbs},
					{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController", Verbs: stdVerbs},
				},
			},
			{
				GroupVersion: extv1beta1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet", Verbs: stdVerbs},
				},
			},
		},
	},
	{
		title:   "Eliminate cohabitations",
		exclude: emptyExclude,
		expect:  []string{"deployment"},
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: appsv1beta2.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: stdVerbs},
				},
			},
			{
				GroupVersion: extv1beta1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: stdVerbs},
				},
			},
		},
	},
	{
		title:   "Eliminate non listable/getable/watchable",
		exclude: emptyExclude,
		expect:  []string{"bar4"},
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: "foo/v42",
				APIResources: []metav1.APIResource{
					{Name: "bar1", Namespaced: true, Kind: "Bar1", Verbs: []string{"get"}},
					{Name: "bar2", Namespaced: true, Kind: "Bar2", Verbs: []string{"get", "list"}},
					{Name: "bar3", Namespaced: true, Kind: "Bar3", Verbs: []string{"watch"}},
					{Name: "bar4", Namespaced: true, Kind: "Bar4", Verbs: stdVerbs},
				},
			},
		},
	},

	{
		title:   "Eliminate unparsable groups",
		exclude: emptyExclude,
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: "foo/bar/baz/v42",
				APIResources: []metav1.APIResource{
					{Name: "spam2", Namespaced: true, Kind: "Spam2", Verbs: stdVerbs},
				},
			},
		},
	},

	{
		title:   "Eliminate subresources",
		exclude: emptyExclude,
		expect:  []string{"pod", "replicationcontroller"},
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: corev1.SchemeGroupVersion.String(),
				APIResources: []metav1.APIResource{
					{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: stdVerbs},
					{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController", Verbs: stdVerbs},
					{Name: "replicationcontrollers/scale", Namespaced: true, Kind: "Scale", Group: "autoscaling", Version: "v1", Verbs: stdVerbs},
				},
			},
		},
	},

	{
		title:   "Eliminate user filtered",
		exclude: []string{"bar1", "bar3"},
		expect:  []string{"bar2", "bar4"},
		resources: []*metav1.APIResourceList{
			{
				GroupVersion: "foo/v42",
				APIResources: []metav1.APIResource{
					{Name: "bar1", Namespaced: true, Kind: "Bar1", Verbs: stdVerbs},
					{Name: "bar2", Namespaced: true, Kind: "Bar2", Verbs: stdVerbs},
					{Name: "bar3", Namespaced: true, Kind: "Bar3", Verbs: stdVerbs},
					{Name: "bar4", Namespaced: true, Kind: "Bar4", Verbs: stdVerbs},
				},
			},
		},
	},
}

func TestObserver(t *testing.T) {
	for _, tt := range resourcesTests {
		factory := new(mockFactory)
		obs := New(new(mockLog), new(mockClient), &mockNotifier{}, factory, tt.exclude)

		client := fakeclientset.NewSimpleClientset()
		fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)
		fakeDiscovery.Resources = tt.resources
		obs.discovery = fakeDiscovery

		obs.Start()
		obs.Stop()

		sort.Strings(factory.names)
		sort.Strings(tt.expect)
		if !reflect.DeepEqual(factory.names, tt.expect) {
			t.Errorf("%s failed: expected %v actual %v", tt.title, tt.expect, factory.names)
		}
	}
}

var duplicatesTest = []*metav1.APIResourceList{
	{

		GroupVersion: corev1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: stdVerbs},
		},
	},
	{
		GroupVersion: extv1beta1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet", Verbs: stdVerbs},
			{Name: "deployments", Namespaced: true, Kind: "Deployment", Verbs: stdVerbs},
		},
	},
}

func TestObserverDuplicas(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = duplicatesTest

	factory := new(mockFactory)
	obs := New(new(mockLog), new(mockClient), &mockNotifier{}, factory, make([]string, 0))
	obs.discovery = fakeDiscovery
	obs.Start()
	err := obs.refresh()
	if err != nil {
		t.Errorf("refresh failed: %v", err)
	}
	obs.Stop()

	expected := []string{"pod", "replicaset", "deployment"}
	sort.Strings(factory.names)
	sort.Strings(expected)
	if !reflect.DeepEqual(factory.names, expected) {
		t.Errorf("%s failed: expected %v actual %v", "Eliminate duplicates", expected, factory.names)
	}
}

func TestObserverRecoverFromDicoveryFailure(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, _ := client.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = duplicatesTest

	fakeClient := &fakerest.RESTClient{
		NegotiatedSerializer: scheme.Codecs,
		Resp: &http.Response{
			StatusCode: http.StatusOK,
		},
	}

	factory := new(mockFactory)
	obs := New(new(mockLog), new(mockClient), &mockNotifier{}, factory, make([]string, 0))

	// failing discovery
	obs.discovery.RESTClient().(*rest.RESTClient).Client = fakeClient.Client
	obs.Start()
	obs.Stop()

	// should resume discovery
	obs.discovery = fakeDiscovery
	obs.Start()
	obs.Stop()

	expected := []string{"pod", "replicaset", "deployment"}
	sort.Strings(factory.names)
	sort.Strings(expected)
	if !reflect.DeepEqual(factory.names, expected) {
		t.Errorf("%s failed: expected %v actual %v", "Recover from failure", expected, factory.names)
	}
}
