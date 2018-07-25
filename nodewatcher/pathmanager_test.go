package main

import (
	"testing"
)

func TestNewPathManager(t *testing.T) {
	pm := NewPathManager(make(chan PathEvent))
	if pm == nil {
		t.Fatal("pm should not be nil")
	}
}

func Test_handleList(t *testing.T) {
	pathCh := make(chan PathEvent)
	pm := NewPathManager(pathCh)

	eventsCh := pm.GetEventsChan()
	go func() {
		eventsCh <- PathEvent{"/my/path", Create}
		eventsCh <- PathEvent{"/your/path", Remove}
	}()
	path := <-pathCh
	if path.Path != "/my/path" {
		t.Fatalf("path.path should be '/my/path', instead is '%s'", path.Path)
	}
	if path.Event != Create {
		t.Fatalf("path.event should be 'Create', instead is '%s'", path.Event)
	}
	path = <-pathCh
	if path.Path != "/your/path" {
		t.Fatalf("path.path should be '/your/path', instead is '%s'", path.Path)
	}
	if path.Event != Remove {
		t.Fatalf("path.event should be 'Remove', instead is '%s'", path.Event)
	}
}

func TestGetEventsChan(t *testing.T) {
	pm := NewPathManager(make(chan PathEvent))
	if pm.GetEventsChan() == nil {
		t.Fatal("GetEventsChan should not return a nil channel")
	}
}
