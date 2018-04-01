package recorder

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controllers"
)

type activeFiles map[string]bool

// Listener receive events from controllers and save them to disk as yaml files
type Listener struct {
	config      *config.KdnConfig
	chans       []chan controllers.Event
	actives     activeFiles
	activesLock sync.RWMutex
}

// New creates a new Listener
func New(config *config.KdnConfig, chans []chan controllers.Event) *Listener {
	return &Listener{
		config:  config,
		chans:   chans,
		actives: activeFiles{},
	}
}

// Watch receive events and persists them to disk
func (w *Listener) Watch() {
	err := os.MkdirAll(filepath.Clean(w.config.LocalDir), 0700)
	if err != nil {
		panic(fmt.Sprintf("Can't create directory %s: %v", w.config.LocalDir, err))
	}

	go w.garbageCollect()

	for {
		w.processNextEvent()
	}
}

func (w *Listener) processNextEvent() {
	cases := make([]reflect.SelectCase, len(w.chans))
	for i, ch := range w.chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	_, value, ok := reflect.Select(cases)
	if !ok {
		return
	}

	ev := value.Interface().(controllers.Event)

	if w.shouldIgnore(ev) {
		return
	}

	w.config.Logger.Debugf("kind=%s name=%s", ev.Kind, ev.Key)

	path, err := getPath(w.config.LocalDir, ev)
	if err != nil {
		w.config.Logger.Errorf("failed to get %s path: %v", ev.Key, err)
	}

	switch ev.Action {
	case controllers.Upsert:
		err = w.save(path, ev.Obj)
	case controllers.Delete:
		err = w.remove(path)
	}

	if err != nil {
		w.config.Logger.Errorf("failed to delete or save %s: %v", ev.Key, err)
	}
}

func (w *Listener) shouldIgnore(ev controllers.Event) bool {
	for _, kind := range w.config.ExcludeKind {
		if strings.Compare(strings.ToLower(kind), ev.Kind) == 0 {
			return true
		}
	}

	for _, obj := range w.config.ExcludeObject {
		if strings.Compare(strings.ToLower(obj), ev.Kind+":"+ev.Key) == 0 {
			return true
		}
	}

	return w.config.DryRun
}

func getPath(root string, ev controllers.Event) (string, error) {
	filename := ev.Kind + "-" + filepath.Base(ev.Key) + ".yaml"

	dir, err := filepath.Abs(filepath.Dir(root + "/" + ev.Key))
	if err != nil {
		return "", err
	}

	return dir + "/" + filename, nil
}

func (w *Listener) remove(file string) error {
	w.activesLock.Lock()
	delete(w.actives, file)
	w.activesLock.Unlock()
	return os.Remove(filepath.Clean(file))
}

func (w *Listener) relativePath(file string) string {
	root := filepath.Clean(w.config.LocalDir)
	return strings.Replace(file, root+"/", "", 1)
}

func (w *Listener) save(file string, data string) error {
	dir := filepath.Clean(filepath.Dir(file))

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("can't create local directory %s: %v", dir, err)
	}

	w.activesLock.Lock()
	w.actives[w.relativePath(file)] = true
	w.activesLock.Unlock()

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create %s on disk: %v", file, err)
	}

	_, err = f.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write to %s on disk: %v", file, err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("failed to close %s file: %v", file, err)
	}

	return nil
}

func (w *Listener) garbageCollect() {
	gcTick := time.NewTicker(w.config.ResyncIntv * 2).C
	for {
		<-gcTick
		w.deleteObsoleteFiles()
	}
}

func (w *Listener) deleteObsoleteFiles() {
	w.activesLock.RLock()
	defer w.activesLock.RUnlock()
	root := filepath.Clean(w.config.LocalDir)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, "yaml") {
			return nil
		}

		_, ok := w.actives[w.relativePath(path)]
		if ok {
			return nil
		}

		return os.Remove(filepath.Clean(path))
	})

	if err != nil {
		w.config.Logger.Warnf("failed to gc some files: %v", err)
	}
}
