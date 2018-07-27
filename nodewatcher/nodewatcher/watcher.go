package nodewatcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Watcher struct {
	dir     string
	watcher *fsnotify.Watcher
	quitCh  chan struct{}
}

func NewWatcher(dir string) *Watcher {
	w := new(Watcher)
	w.dir = dir
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Error(err)
		return nil
	}
	w.watcher = watcher
	w.quitCh = make(chan struct{})
	return w
}

func (w *Watcher) CheckDir() error {
	info, err := os.Lstat(w.dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", w.dir)
	}
	err = w.watcher.Add(w.dir)
	if err != nil {
		return fmt.Errorf("'%s' : %s", w.dir, err)
	}
	return nil
}

// WatchDir recursively watches all files in `dir` directory and its subdirectories.
func (w *Watcher) WatchDir(pathCh chan<- []shared.Operation) error {
	operations := []shared.Operation{}
	err := filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		select {
		case <-w.quitCh:
			return shared.ErrQuit
		default:
		}
		if err != nil {
			glog.Error(err)
			return err
		}
		newPath, err := Chroot(path, w.dir)
		if err != nil {
			glog.Error(err)
			return err
		}
		operations = append(operations, shared.Operation{Path: newPath, Event: shared.Create})
		if info.IsDir() {
			if err := w.watcher.Add(path); err != nil {
				glog.Error(err)
				return err
			}
		}
		return nil
	})
	pathCh <- operations
	return err
}

// HandleFileEvents notifies our pathmanager whenever there are new files or deleted files in the directory watched
// by the `Watcher` as well as its subdirectories.
func (w *Watcher) HandleFileEvents(pathCh chan<- []shared.Operation) {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				newPath, err := Chroot(event.Name, w.dir)
				if err != nil {
					glog.Error(err)
					continue
				}
				pathCh <- []shared.Operation{shared.Operation{Path: newPath, Event: shared.Create}}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {
				newPath, err := Chroot(event.Name, w.dir)
				if err != nil {
					glog.Error(err)
					continue
				}
				pathCh <- []shared.Operation{shared.Operation{Path: newPath, Event: shared.Remove}}
			}
		case <-w.quitCh:
			return
		}
	}
}

func (w *Watcher) Stop() {
	close(w.quitCh)
	w.watcher.Close()
}
