package controller

import (
	"strings"
	"testing"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/event"
	"github.com/bpineau/katafygio/pkg/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakecontroller "k8s.io/client-go/tools/cache/testing"
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
				"resourceVersion": 1,
				"uid":             "00000000-0000-0000-0000-000000000042",
				"selfLink":        "shouldnotbethere",
			},
			"status": "shouldnotbethere",
		},
	}
)

func TestController(t *testing.T) {

	conf := &config.KfConfig{
		Logger:        log.New("info", "", "test"),
		ExcludeObject: []string{"pod:ns3/Bar3"},

		// label filters can't be tested due to the way we inject objets in tests
		Filter: "label1=something",
	}

	client := fakecontroller.NewFakeControllerSource()

	evt := new(mockNotifier)
	f := new(Factory)
	ctrl := f.NewController(client, evt, "pod", conf)

	// this will trigger a deletion event
	idx := ctrl.(*Controller).informer.GetIndexer()
	err := idx.Add(obj1)
	if err != nil {
		t.Errorf("failed to inject an object in indexer: %v", err)
	}

	client.Add(obj2)
	client.Add(obj3)

	ctrl.Start()
	// wait until queue is drained
	for ctrl.(*Controller).queue.Len() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	ctrl.Stop()

	gotFoo2 := false
	for _, ev := range evt.evts {
		// ensure cleanup filters works as expected
		if strings.Contains(ev.Object, "shouldnotbethere") {
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
	}

	if !gotFoo2 {
		t.Errorf("we should have notified obj2")
	}
}
