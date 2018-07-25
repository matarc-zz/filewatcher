package main

import (
	"encoding/gob"
	"net"
	"testing"
)

func TestRun(t *testing.T) {
	clt := NewClient("localhost:12345")
	pathCh := make(chan PathEvent)
	listener, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		t.Fatal(err)
	}
	go clt.Run(pathCh)
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	dec := gob.NewDecoder(conn)
	pathCh <- PathEvent{"/my/path", Create}
	var buf PathEvent
	err = dec.Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Path != "/my/path" {
		t.Fatalf("buf.path should be '/my/path', instead is '%s'", buf.Path)
	}
	if buf.Event != Create {
		t.Fatalf("buf.event should be 'Create', instead is '%s'", buf.Event)
	}
}

func TestNewClient(t *testing.T) {
	clt := NewClient("localhost:12345")
	if clt == nil {
		t.Fatalf("NewClient should not return a nil value")
	}
	if !clt.bufIsEmpty {
		t.Fatalf("bufIsEmpty initial value should be 'true'")
	}
	if clt.quitCh == nil {
		t.Fatalf("quitCh should not be a nil channel")
	}
	if clt.serverAddress != "localhost:12345" {
		t.Fatalf("serverAddress should be 'localhost:12345', instead is '%s'", clt.serverAddress)
	}
}
