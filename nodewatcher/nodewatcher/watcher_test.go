package nodewatcher

import (
	"io/ioutil"
	"os"
	"path"
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
	pathCh := make(chan []shared.Operation)
	go func() {
		<-pathCh
	}()
	err = w.WatchDir(pathCh)
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

func TestWatchDirListFiles(t *testing.T) {
	files := make(map[string]bool)
	tmpDir, err := ioutil.TempDir("", "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	files[path.Base(tmpDir)] = true
	// Create a file in tmpDir
	tmpFile, err := ioutil.TempFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err := Chroot(tmpFile.Name(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	// Create a subdirectory in tmpDir
	subTmpDir, err := ioutil.TempDir(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err = Chroot(subTmpDir, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	// Create a file in that subdirectory
	tmpFile, err = ioutil.TempFile(subTmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err = Chroot(tmpFile.Name(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	pathCh := make(chan []shared.Operation)
	w := NewWatcher(tmpDir)
	go w.WatchDir(pathCh)
	paths := <-pathCh
	for _, path := range paths {
		if _, ok := files[path.Path]; !ok {
			t.Fatalf("`%s` should be in the list of files : `%v`", path, files)
		}
	}
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
	pathCh := make(chan []shared.Operation)
	go func() {
		<-pathCh
	}()
	err = w.WatchDir(pathCh)
	if err != nil {
		t.Fatal(err)
	}
	go w.HandleFileEvents(pathCh)

	// Create a file in subDir
	file, err := ioutil.TempFile(subDir, "")
	if err != nil {
		t.Fatal(err)
	}
	pathEvent := <-pathCh
	if pathEvent[0].Event&shared.Create != shared.Create {
		t.Fatalf("pathEvent.Event should be 'Create', instead is '%s'", pathEvent[0].Event)
	}
	path, err := Chroot(file.Name(), rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent[0].Path != path {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", path, pathEvent[0].Path)
	}

	// Rename a file from subdir into rootDir
	newPath := filepath.Join(rootDir, filepath.Base(file.Name()))
	err = os.Rename(file.Name(), newPath)
	if err != nil {
		t.Fatal(err)
	}
	pathEvent = <-pathCh
	if pathEvent[0].Event&shared.Create != shared.Create {
		t.Fatalf("pathEvent.Event should be 'Create', instead is '%s'", pathEvent[0].Event)
	}
	newPath, err = Chroot(newPath, rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent[0].Path != newPath {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", newPath, pathEvent[0].Path)
	}
	pathEvent = <-pathCh
	if pathEvent[0].Event&shared.Remove != shared.Remove {
		t.Fatalf("pathEvent.Event should be 'Remove', instead is '%s'", pathEvent[0].Event)
	}
	if pathEvent[0].Path != path {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", path, pathEvent[0].Path)
	}

	// Remove a file from rootDir
	newPath = filepath.Join(rootDir, filepath.Base(file.Name()))
	err = os.Remove(newPath)
	if err != nil {
		t.Fatal(err)
	}
	pathEvent = <-pathCh
	if pathEvent[0].Event&shared.Remove != shared.Remove {
		t.Fatalf("pathEvent.Event should be 'Remove', instead is '%s'", pathEvent[0].Event)
	}
	newPath, err = Chroot(newPath, rootDir)
	if err != nil {
		t.Fatal(err)
	}
	if pathEvent[0].Path != newPath {
		t.Fatalf("pathEvent.Path should be '%s', instead is '%s'", newPath, pathEvent[0].Path)
	}
}

func TestCheckDir(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filepath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	myDir := filepath.Join(rootDir, "my")
	err = os.Mkdir(myDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	w := NewWatcher(myDir)
	err = w.CheckDir()
	if err != nil {
		t.Fatal(err)
	}
	w.Stop()

	err = os.Chmod(myDir, 0000)
	if err != nil {
		t.Fatal(err)
	}
	w = NewWatcher(myDir)
	err = w.CheckDir()
	if err == nil {
		t.Fatalf("CheckDir should return an error on a directory that can't be read")
	}
	w.Stop()

	file, err := ioutil.TempFile(rootDir, "")
	if err != nil {
		t.Fatal(err)
	}
	w = NewWatcher(file.Name())
	err = w.CheckDir()
	if err == nil {
		t.Fatalf("CheckDir should return an error on a file")
	}
}
