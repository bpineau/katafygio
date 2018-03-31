package recorder

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bpineau/katafygio/config"
	"github.com/bpineau/katafygio/pkg/controllers"
)

// Watch receive events and persists them to disk
func Watch(config *config.KdnConfig, chans []chan controllers.Event) {
	err := os.MkdirAll(filepath.Clean(config.LocalDir), 0700)
	if err != nil {
		panic(fmt.Sprintf("Can't create local directory %s: %v", config.LocalDir, err))
	}

	for {
		processNextEvent(config, chans)
	}
}

func processNextEvent(config *config.KdnConfig, chans []chan controllers.Event) {
	cases := make([]reflect.SelectCase, len(chans))
	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}
	_, value, ok := reflect.Select(cases)
	if !ok {
		return
	}

	ev := value.Interface().(controllers.Event)

	if shouldIgnore(config, ev) {
		return
	}

	config.Logger.Debugf("kind=%s name=%s", ev.Kind, ev.Key)

	path, err := getPath(config.LocalDir, ev)
	if err != nil {
		config.Logger.Errorf("failed to get %s path: %v", ev.Key, err)
	}

	switch ev.Action {
	case controllers.Upsert:
		err = save(path, ev.Obj)
	case controllers.Delete:
		err = remove(path)
	}

	if err != nil {
		config.Logger.Errorf("failed to delete or save %s: %v", ev.Key, err)
	}
}

func shouldIgnore(config *config.KdnConfig, ev controllers.Event) bool {
	for _, kind := range config.ExcludeKind {
		if strings.Compare(strings.ToLower(kind), ev.Kind) == 0 {
			return true
		}
	}

	for _, obj := range config.ExcludeObject {
		if strings.Compare(strings.ToLower(obj), ev.Kind+":"+ev.Key) == 0 {
			return true
		}
	}

	return config.DryRun
}

func getPath(root string, ev controllers.Event) (string, error) {
	filename := ev.Kind + "-" + filepath.Base(ev.Key) + ".yaml"

	dir, err := filepath.Abs(filepath.Dir(root + "/" + ev.Key))
	if err != nil {
		return "", err
	}

	return dir + "/" + filename, nil
}

func remove(file string) error {
	return os.Remove(filepath.Clean(file))
}

func save(file string, data string) error {
	dir := filepath.Clean(filepath.Dir(file))

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("can't create local directory %s: %v", dir, err)
	}

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
