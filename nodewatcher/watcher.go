package main

import (
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

// WatchDir recursively watches all files in `dir` directory and its subdirectories.
func (w *Watcher) WatchDir() error {
	return filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			glog.Error(err)
			return err
		}
		if info.IsDir() {
			if err := w.watcher.Add(path); err != nil {
				glog.Error(err)
				return err
			}
		}
		return nil
	})
}

// HandleFileEvents notifies our pathmanager whenever there are new files or deleted files in the directory watched
// by the `Watcher` as well as its subdirectories.
func (w *Watcher) HandleFileEvents(pathCh chan<- shared.Operation) {
	for {
		select {
		case event := <-w.watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				newPath, err := Chroot(event.Name, w.dir)
				if err != nil {
					glog.Error(err)
					continue
				}
				pathCh <- shared.Operation{Path: newPath, Event: shared.Create}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {
				newPath, err := Chroot(event.Name, w.dir)
				if err != nil {
					glog.Error(err)
					continue
				}
				pathCh <- shared.Operation{Path: newPath, Event: shared.Remove}
			}
		}
	}
}
