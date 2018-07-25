package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchDir(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	successCh := make(chan bool)
	quitCh := make(chan struct{})
	go func() {
		for {
			select {
			case event := <-watcher.Events:
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
	err = WatchDir(rootDir, watcher)
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
