package event

import (
	"reflect"
	"testing"
)

var (
	notif = Notification{
		Action: Upsert,
		Key:    "foo",
		Kind:   "bar",
		Object: []byte("spam egg"),
	}
)

func TestEvent(t *testing.T) {
	ev := New()

	go ev.Send(&notif)

	reader := ev.ReadChan()
	got := <-reader

	if !reflect.DeepEqual(notif, got) {
		t.Errorf("notification failed: expected %v actual %v", notif, got)
	}
}
