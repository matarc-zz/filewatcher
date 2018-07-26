package storage

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"
)

func TestNewServer(t *testing.T) {
	srv := NewServer("localhost:12345", "/my/path")
	if srv == nil {
		t.Fatalf("NewServer should not return a nil value")
	}
	if srv.Address != "localhost:12345" {
		t.Fatalf("srv.address should be 'localhost:12345', instead is '%s'", srv.Address)
	}
	if srv.quitCh == nil {
		t.Fatalf("srv.quitCh should not be nil")
	}
	if srv.DbPath != "/my/path" {
		t.Fatalf("srv.dbPath should be '/my/path', instead is '%s'", srv.DbPath)
	}
}

func TestRun(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	dbPath := filepath.Join(rootDir, "mydb")
	srv := NewServer("localhost:12345", dbPath)
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run()
	}()
	clt, err := rpc.Dial("tcp", "localhost:12345")
	if err != nil {
		t.Fatal(err)
	}
	defer clt.Close()
	srv.Stop()
	<-done
	_, err = rpc.Dial("tcp", "localhost:12345")
	if err == nil {
		t.Fatalf("Listener should no longer be listening")
	}
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		t.Fatalf("'%s' was not created", dbPath)
	}
}
