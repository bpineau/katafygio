package recorder

import (
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/bpineau/katafygio/pkg/event"
)

type mockLog struct{}

func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Errorf(format string, args ...interface{}) {}

func newNotif(action event.Action, key string) *event.Notification {
	return &event.Notification{
		Action: action,
		Key:    key,
		Kind:   "foo",
		Object: []byte("bar"),
	}
}

var (
	logs    = new(mockLog)
	fakedir = "/tmp/ktest"
)

func TestRecorder(t *testing.T) {
	appFs = afero.NewMemMapFs()

	evt := event.New()

	rec := New(logs, evt, fakedir, 120, false).Start()

	evt.Send(newNotif(event.Upsert, "foo1"))
	evt.Send(newNotif(event.Upsert, "foo2"))
	evt.Send(newNotif(event.Delete, "foo1"))
	evt.Send(newNotif(event.Upsert, "foo2"))

	rec.Stop() // to flush ongoing fs operations

	exist, _ := afero.Exists(appFs, fakedir+"/foo-foo2.yaml")
	if !exist {
		t.Error("foo-foo2.yaml should exist; upsert event didn't propagate")
	}

	exist, _ = afero.Exists(appFs, fakedir+"/foo-foo1.yaml")
	if exist {
		t.Error("foo-foo1.yaml shouldn't exist, delete event didn't propagate")
	}

	rogue := fakedir + "/roguefile.yaml"
	_ = afero.WriteFile(appFs, rogue, []byte{42}, 0600)
	_ = afero.WriteFile(appFs, rogue+".txt", []byte{42}, 0600)
	rec.deleteObsoleteFiles()

	exist, _ = afero.Exists(appFs, rogue)
	if exist {
		t.Errorf("%s file should have been garbage collected", rogue)
	}

	exist, _ = afero.Exists(appFs, rogue+".txt")
	if !exist {
		t.Errorf("garbage collection should only touch .yaml files")
	}
}

func TestDryRunRecorder(t *testing.T) {
	appFs = afero.NewMemMapFs()

	dryevt := event.New()
	dryrec := New(logs, dryevt, fakedir, 60, true).Start()
	dryevt.Send(newNotif(event.Upsert, "foo3"))
	dryevt.Send(newNotif(event.Upsert, "foo4"))
	dryevt.Send(newNotif(event.Delete, "foo4"))
	dryrec.Stop()

	exist, _ := afero.Exists(appFs, fakedir+"/foo-foo3.yaml")
	if exist {
		t.Error("foo-foo3.yaml was created but we're in dry-run mode")
	}

	rogue := fakedir + "/roguefile.yaml"
	_ = afero.WriteFile(appFs, rogue, []byte{42}, 0600)
	dryrec.deleteObsoleteFiles()

	exist, _ = afero.Exists(appFs, rogue)
	if !exist {
		t.Errorf("garbage collection shouldn't remove files in dry-run mode")
	}
}

// testing behavior on fs errors (we shouldn't block the program)
func TestFailingFSRecorder(t *testing.T) {
	appFs = afero.NewMemMapFs()

	evt := event.New()

	rec := New(logs, evt, fakedir, 60, false).Start()

	_ = afero.WriteFile(appFs, fakedir+"/foo.yaml", []byte{42}, 0600)

	// switching to failing (read-only) filesystem
	appFs = afero.NewReadOnlyFs(appFs)

	err := rec.save("foo", []byte("bar"))
	if err == nil {
		t.Error("save should return an error in case of failure")
	}

	// shouldn't panic in case of failures
	rec.deleteObsoleteFiles()

	// shouldn't block (the controllers event loop will retry anyway)
	ch := make(chan struct{})
	go func() {
		evt.Send(newNotif(event.Upsert, "foo3"))
		evt.Send(newNotif(event.Upsert, "foo4"))
		ch <- struct{}{}
	}()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Error("recorder shouldn't block in case of fs failure")
	}

	// back to normal operations
	rec.Stop() // just to flush ongoing ops before switch filesystem
	appFs = afero.NewMemMapFs()
	rec.stopch = make(chan struct{})
	rec.donech = make(chan struct{})
	rec.Start()
	evt.Send(newNotif(event.Upsert, "foo2"))
	rec.Stop() // flush ongoing ops

	exist, _ := afero.Exists(appFs, fakedir+"/foo-foo2.yaml")
	if !exist {
		t.Error("foo-foo2.yaml should exist; recorder should recover from fs failures")
	}
}
