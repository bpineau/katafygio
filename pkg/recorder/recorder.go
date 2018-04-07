package recorder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controller"
)

type activeFiles map[string]bool

// Listener receive events from controllers and save them to disk as yaml files
type Listener struct {
	config      *config.KfConfig
	evchan      chan controller.Event
	actives     activeFiles
	activesLock sync.RWMutex
	stopch      chan struct{}
	donech      chan struct{}
}

// New creates a new Listener
func New(config *config.KfConfig, evchan chan controller.Event) *Listener {
	return &Listener{
		config:  config,
		evchan:  evchan,
		actives: activeFiles{},
	}
}

// Start receive events and persists them to disk as files
func (w *Listener) Start() *Listener {
	w.config.Logger.Info("Starting event recorder")
	err := os.MkdirAll(filepath.Clean(w.config.LocalDir), 0700)
	if err != nil {
		panic(fmt.Sprintf("Can't create directory %s: %v", w.config.LocalDir, err))
	}

	go func() {
		gcTick := time.NewTicker(w.config.ResyncIntv * 2)
		w.stopch = make(chan struct{})
		w.donech = make(chan struct{})
		defer gcTick.Stop()
		defer close(w.donech)

		for {
			select {
			case <-w.stopch:
				return
			case ev := <-w.evchan:
				w.processNextEvent(ev)
			case <-gcTick.C:
				w.deleteObsoleteFiles()
			}
		}
	}()

	return w
}

// Stop halts the recorder service
func (w *Listener) Stop() {
	w.config.Logger.Info("Stopping event recorder")
	close(w.stopch)
	<-w.donech
}

func (w *Listener) processNextEvent(ev controller.Event) {
	if w.shouldIgnore(ev) {
		return
	}

	w.config.Logger.Debugf("kind=%s name=%s", ev.Kind, ev.Key)

	path, err := getPath(w.config.LocalDir, ev)
	if err != nil {
		w.config.Logger.Errorf("failed to get %s path: %v", ev.Key, err)
	}

	switch ev.Action {
	case controller.Upsert:
		err = w.save(path, ev.Obj)
	case controller.Delete:
		err = w.remove(path)
	}

	if err != nil {
		w.config.Logger.Errorf("failed to delete or save %s: %v", ev.Key, err)
	}
}

func (w *Listener) shouldIgnore(ev controller.Event) bool {
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

func getPath(root string, ev controller.Event) (string, error) {
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
