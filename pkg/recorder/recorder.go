// Package recorder listen for event notifications from controllers,
// and persists those events' content as files on disk.
package recorder

import (
	"fmt"
	"hash/crc64"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/event"
)

var (
	appFs      = afero.NewOsFs()
	crc64Table = crc64.MakeTable(crc64.ECMA)
)

// activeFiles will contain a list of active (present in cluster) objets; we'll
// use that to periodically find and garbage collect stale objets in the git repos
// (ie. if some objects were delete from cluster while katafygio was not running),
// and to skip already existing and unchanged files.
type activeFiles map[string]uint64

// Listener receive events from controllers and save them to disk as yaml files
type Listener struct {
	config      *config.KfConfig
	events      event.Notifier
	actives     activeFiles
	activesLock sync.RWMutex
	stopch      chan struct{}
	donech      chan struct{}
}

// New creates a new event Listener
func New(config *config.KfConfig, events event.Notifier) *Listener {
	return &Listener{
		config:  config,
		events:  events,
		actives: activeFiles{},
	}
}

// Start continuously receive events and saves them to disk as files
func (w *Listener) Start() *Listener {
	w.config.Logger.Info("Starting event recorder")
	err := appFs.MkdirAll(filepath.Clean(w.config.LocalDir), 0700)
	if err != nil {
		panic(fmt.Sprintf("Can't create directory %s: %v", w.config.LocalDir, err))
	}

	go func() {
		evCh := w.events.ReadChan()
		gcTick := time.NewTicker(w.config.ResyncIntv * 2)
		w.stopch = make(chan struct{})
		w.donech = make(chan struct{})
		defer gcTick.Stop()
		defer close(w.donech)

		for {
			select {
			case <-w.stopch:
				return
			case ev := <-evCh:
				w.processNextEvent(&ev)
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

func (w *Listener) processNextEvent(ev *event.Notification) {
	path, err := getPath(w.config.LocalDir, ev)
	if err != nil {
		w.config.Logger.Errorf("failed to get %s path: %v", ev.Key, err)
	}

	switch ev.Action {
	case event.Upsert:
		err = w.save(path, ev.Object)
	case event.Delete:
		err = w.remove(path)
	}

	if err != nil {
		w.config.Logger.Errorf("failed to delete or save %s: %v", ev.Key, err)
	}
}

func getPath(root string, ev *event.Notification) (string, error) {
	filename := ev.Kind + "-" + filepath.Base(ev.Key) + ".yaml"

	dir, err := filepath.Abs(filepath.Dir(root + "/" + ev.Key))
	if err != nil {
		return "", err
	}

	return dir + "/" + filename, nil
}

func (w *Listener) remove(file string) error {
	w.config.Logger.Debugf("Removing %s from disk", file)
	if w.config.DryRun {
		return nil
	}

	w.activesLock.Lock()
	delete(w.actives, file)
	w.activesLock.Unlock()
	return appFs.Remove(filepath.Clean(file))
}

func (w *Listener) relativePath(file string) string {
	root := filepath.Clean(w.config.LocalDir)
	return strings.Replace(file, root+"/", "", 1)
}

func (w *Listener) save(file string, data []byte) error {
	w.config.Logger.Debugf("Saving %s to disk", file)

	if w.config.DryRun {
		return nil
	}

	csum := crc64.Checksum(data, crc64Table)

	w.activesLock.RLock()
	prevsum, ok := w.actives[w.relativePath(file)]
	w.activesLock.RUnlock()
	if ok && prevsum == csum {
		return nil
	}

	dir := filepath.Clean(filepath.Dir(file))

	err := appFs.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("can't create local directory %s: %v", dir, err)
	}

	tmpf, err := afero.TempFile(appFs, "", "katafygio")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file: %v", err)
	}

	_, err = tmpf.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to %s on disk: %v", tmpf.Name(), err)
	}

	if err := tmpf.Close(); err != nil {
		return fmt.Errorf("failed to close a temporary file: %v", err)
	}

	if err := appFs.Rename(tmpf.Name(), file); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %v", tmpf.Name(), file, err)
	}

	w.activesLock.Lock()
	w.actives[w.relativePath(file)] = csum
	w.activesLock.Unlock()

	return nil
}

func (w *Listener) deleteObsoleteFiles() {
	w.activesLock.RLock()
	defer w.activesLock.RUnlock()
	root := filepath.Clean(w.config.LocalDir)

	err := afero.Walk(appFs, root, func(path string, info os.FileInfo, err error) error {
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

		w.config.Logger.Debugf("Removing %s from disk", path)
		if !w.config.DryRun {
			return appFs.Remove(filepath.Clean(path))
		}

		return nil
	})

	if err != nil {
		w.config.Logger.Warnf("failed to gc some files: %v", err)
	}
}
