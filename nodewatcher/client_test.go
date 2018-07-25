package main

import (
	"encoding/gob"
	"net"
	"testing"

	"github.com/matarc/filewatcher/shared"
)

func TestRun(t *testing.T) {
	clt := NewClient("localhost:12345", "client1")
	pathCh := make(chan shared.PathEvent)
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
	pathCh <- shared.PathEvent{"/my/path", shared.Create}
	var packet shared.Packet
	err = dec.Decode(&packet)
	if err != nil {
		t.Fatal(err)
	}
	if packet.PathEvent.Path != "/my/path" {
		t.Fatalf("Path should be '/my/path', instead is '%s'", packet.PathEvent.Path)
	}
	if packet.PathEvent.Event != shared.Create {
		t.Fatalf("Event should be 'Create', instead is '%s'", packet.PathEvent.Event)
	}
	if packet.Id != "client1" {
		t.Fatalf("Id should be 'client1', instead is '%s'", packet.Id)
	}
}

func TestNewClient(t *testing.T) {
	clt := NewClient("localhost:12345", "client1")
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
	if clt.id != "client1" {
		t.Fatalf("id should be 'client1', instead is '%s'", clt.id)
	}
}
