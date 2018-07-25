package main

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
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
