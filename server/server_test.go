package main

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	srv := NewServer("localhost:12345", "/my/path")
	if srv == nil {
		t.Fatalf("NewServer should not return a nil value")
	}
	if srv.address != "localhost:12345" {
		t.Fatalf("srv.address should be 'localhost:12345', instead is '%s'", srv.address)
	}
	if srv.quitCh == nil {
		t.Fatalf("srv.quitCh should not be nil")
	}
	if srv.dbPath != "/my/path" {
		t.Fatalf("srv.dbPath should be '/my/path', instead is '%s'", srv.dbPath)
	}
}
