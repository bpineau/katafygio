package controller

import (
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/katafygio/pkg/event"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakecontroller "k8s.io/client-go/tools/cache/testing"
	"k8s.io/klog"
)

type mockNotifier struct {
	evts []*event.Notification
}

func (m *mockNotifier) Send(ev *event.Notification) {
	m.evts = append(m.evts, ev)
}

func (m *mockNotifier) ReadChan() <-chan event.Notification {
	return make(chan event.Notification)
}

type mockLog struct{}

func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Errorf(format string, args ...interface{}) {}

var (
	obj1 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo1",
			"metadata": map[string]interface{}{
				"name":            "Bar1",
				"namespace":       "ns1",
				"resourceVersion": 1,
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}

	obj2 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo2",
			"metadata": map[string]interface{}{
				"name":            "Bar2",
				"namespace":       "ns2",
				"resourceVersion": 1,
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}

	obj3 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo3",
			"metadata": map[string]interface{}{
				"name":            "Bar3",
				"namespace":       "ns3",
				"resourceVersion": "1",
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}

	obj4 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo2",
			"metadata": map[string]interface{}{
				"name":            "Bar4",
				"namespace":       "ns4",
				"resourceVersion": "2",
				"foo":             "canary-bar4",
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}

	obj5 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo2",
			"metadata": map[string]interface{}{
				"name":            "Bar4",
				"namespace":       "ns4",
				"ResourceVersion": "4",
				"foo":             "canary-bar5",
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}
)

func init() {
	// Enable klog which is used in dependencies
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")
	_ = flag.Set("v", "9")
}

func TestController(t *testing.T) {
	client := fakecontroller.NewFakeControllerSource()

	evt := new(mockNotifier)
	log := new(mockLog)
	f := NewFactory(log, "label1=something", 60, []string{"pod:ns3/Bar3"}, []string{})
	ctrl := f.NewController(client, evt, "pod")

	// this will trigger a deletion event
	idx := ctrl.(*Controller).informer.GetIndexer()
	err := idx.Add(obj1)
	if err != nil {
		t.Errorf("failed to inject an object in indexer: %v", err)
	}

	client.Add(obj2)
	client.Add(obj3)
	client.Add(obj4)
	client.Modify(obj5)

	ctrl.Start()
	// wait until queue is drained
	for ctrl.(*Controller).queue.Len() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	ctrl.Stop()

	gotFoo2 := false
	for _, ev := range evt.evts {
		// ensure cleanup filters works as expected
		if strings.Contains(string(ev.Object), "shouldnotbethere") {
			t.Error("object cleanup filters didn't work")
		}

		// ensure deletion notifications pops up as expected
		if strings.Compare(ev.Key, "ns1/Bar1") == 0 && ev.Action != event.Delete {
			t.Error("deletion notification failed")
		}

		if strings.Compare(ev.Key, "ns2/Bar2") == 0 {
			gotFoo2 = true
		}

		// ensure objet filter works as expected
		if strings.Compare(ev.Key, "ns3/Bar3") == 0 {
			t.Error("execludedobject filter failed")
		}

		// ensure updates propagate
		if strings.Contains(string(ev.Object), "canary-bar4") {
			t.Error("update didn't propagate")
		}
	}

	if !gotFoo2 {
		t.Errorf("we should have notified obj2")
	}
}
