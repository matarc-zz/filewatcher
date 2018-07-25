package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
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
