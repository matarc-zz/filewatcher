package main

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
)

// WatchDir recursively watches all files in `dir` directory and its subdirectories.
func WatchDir(dir string, watcher *fsnotify.Watcher) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			glog.Error(err)
			return err
		}
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				glog.Error(err)
				return err
			}
		}
		return nil
	})
}
