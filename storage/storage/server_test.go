package storage

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"

	"github.com/matarc/filewatcher/shared"
)

func TestInit(t *testing.T) {
	srv := new(Server)
	srv.Init()
	if srv.Address != shared.DefaultStorageAddress {
		t.Fatalf("srv.address should be '%s', instead is '%s'", shared.DefaultStorageAddress, srv.Address)
	}
	if srv.DbPath != shared.DefaultDbPath {
		t.Fatalf("srv.dbPath should be '%s', instead is '%s'", shared.DefaultDbPath, srv.DbPath)
	}

	srv = new(Server)
	srv.Address = "localhost:12345"
	srv.DbPath = "mydb"
	srv.Init()
	if srv.Address != "localhost:12345" {
		t.Fatalf("srv.address should be 'localhost:12345', instead is '%s'", srv.Address)
	}
	if srv.DbPath != "mydb" {
		t.Fatalf("srv.dbPath should be 'mydb', instead is '%s'", srv.DbPath)
	}
}

func TestRun(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	dbPath := filepath.Join(rootDir, "mydb")
	srv := new(Server)
	shared.LoadConfig("", srv)
	srv.DbPath = dbPath
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run()
	}()
	clt, err := rpc.Dial("tcp", shared.DefaultStorageAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer clt.Close()
	srv.Stop()
	<-done
	_, err = rpc.Dial("tcp", shared.DefaultStorageAddress)
	if err == nil {
		t.Fatalf("Listener should no longer be listening")
	}
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		t.Fatalf("'%s' was not created", dbPath)
	}
}
