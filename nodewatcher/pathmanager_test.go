package main

import (
	"testing"

	"github.com/matarc/filewatcher/shared"
)

func TestNewPathManager(t *testing.T) {
	pm := NewPathManager(make(chan shared.PathEvent))
	if pm == nil {
		t.Fatal("pm should not be nil")
	}
}

func Test_handleList(t *testing.T) {
	pathCh := make(chan shared.PathEvent)
	pm := NewPathManager(pathCh)

	eventsCh := pm.GetEventsChan()
	go func() {
		eventsCh <- shared.PathEvent{"/my/path", shared.Create}
		eventsCh <- shared.PathEvent{"/your/path", shared.Remove}
	}()
	path := <-pathCh
	if path.Path != "/my/path" {
		t.Fatalf("path.path should be '/my/path', instead is '%s'", path.Path)
	}
	if path.Event != shared.Create {
		t.Fatalf("path.event should be 'Create', instead is '%s'", path.Event)
	}
	path = <-pathCh
	if path.Path != "/your/path" {
		t.Fatalf("path.path should be '/your/path', instead is '%s'", path.Path)
	}
	if path.Event != shared.Remove {
		t.Fatalf("path.event should be 'Remove', instead is '%s'", path.Event)
	}
}

func TestGetEventsChan(t *testing.T) {
	pm := NewPathManager(make(chan shared.PathEvent))
	if pm.GetEventsChan() == nil {
		t.Fatal("GetEventsChan should not return a nil channel")
	}
}
