package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/matarc/filewatcher/shared"
)

func TestWatchDir(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	tmpDir, err := ioutil.TempDir(rootDir, "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	tmpDir, err = ioutil.TempDir(tmpDir, "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	w := NewWatcher(rootDir)
	if w == nil {
		t.Fatalf("NewWatcher failed to initialize fsnotify.Watcher")
	}
	successCh := make(chan bool)
	quitCh := make(chan struct{})
	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					successCh <- true
				}
			case <-time.After(time.Second):
				successCh <- false
			case <-quitCh:
				return
			}
		}
	}()
	err = w.WatchDir()
	if err != nil {
		t.Fatal(err)
	}
	_, err = ioutil.TempFile(tmpDir, "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	success := <-successCh
	if !success {
		t.Fatalf("Watcher timed out")
	}
	close(quitCh)
}

func TestNewWatcher(t *testing.T) {
	w := NewWatcher("/my/path")
	if w == nil {
		t.Fatalf("NewWatcher failed to initialize fsnotify.Watcher")
	}
	if w.dir != "/my/path" {
		t.Fatalf("w.dir should be '/my/path', instead is '%s'", w.dir)
	}
	if w.watcher == nil {
		t.Fatalf("w.watcher should not be nil")
	}
	if w.quitCh == nil {
		t.Fatalf("w.quitCh should not be nil")
	}
}

func TestHandleFileEvents(t *testing.T) {
	// Setup
	rootDir, err := ioutil.TempDir("", "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	subDir, err := ioutil.TempDir(rootDir, "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	w := NewWatcher(rootDir)
	if w == nil {
		t.Fatalf("NewWatcher failed to initialize fsnotify.Watcher")
	}
	err = w.WatchDir()
	if err != nil {
		t.Fatal(err)
	}
	pathCh := make(chan shared.Operation)
	go w.HandleFileEvents(pathCh)

	// Create a file in subDir
	file, err := ioutil.TempFile(subDir, "")
	if err != nil {
		t.Fatal(err)
	}
	pathEvent := <-pathCh
	if pathEvent.Event&shared.Create != shared.Create {
		t.Fatalf("pathEvent.Event should be 'Create', instead is '%s'", pathEvent.Event)
	}
	path, err := Chroot(file.Name(), rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent.Path != path {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", path, pathEvent.Path)
	}

	// Rename a file from subdir into rootDir
	newPath := filepath.Join(rootDir, filepath.Base(file.Name()))
	err = os.Rename(file.Name(), newPath)
	if err != nil {
		t.Fatal(err)
	}
	pathEvent = <-pathCh
	if pathEvent.Event&shared.Create != shared.Create {
		t.Fatalf("pathEvent.Event should be 'Create', instead is '%s'", pathEvent.Event)
	}
	newPath, err = Chroot(newPath, rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent.Path != newPath {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", newPath, pathEvent.Path)
	}
	pathEvent = <-pathCh
	if pathEvent.Event&shared.Remove != shared.Remove {
		t.Fatalf("pathEvent.Event should be 'Remove', instead is '%s'", pathEvent.Event)
	}
	if pathEvent.Path != path {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", path, pathEvent.Path)
	}

	// Remove a file from rootDir
	newPath = filepath.Join(rootDir, filepath.Base(file.Name()))
	err = os.Remove(newPath)
	if err != nil {
		t.Fatal(err)
	}
	pathEvent = <-pathCh
	if pathEvent.Event&shared.Remove != shared.Remove {
		t.Fatalf("pathEvent.Event should be 'Remove', instead is '%s'", pathEvent.Event)
	}
	newPath, err = Chroot(newPath, rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent.Path != newPath {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", newPath, pathEvent.Path)
	}
}
