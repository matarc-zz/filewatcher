package storage

import (
	"encoding/json"
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"

	"github.com/matarc/filewatcher/shared"
)

func TestLoadConfig(t *testing.T) {
	srv := LoadConfig("")
	if srv == nil {
		t.Fatalf("LoadConfig should not return a nil value")
	}
	if srv.Address != shared.DefaultStorageAddress {
		t.Fatalf("srv.address should be '%s', instead is '%s'", shared.DefaultStorageAddress, srv.Address)
	}
	if srv.DbPath != shared.DefaultDbPath {
		t.Fatalf("srv.dbPath should be '%s', instead is '%s'", shared.DefaultDbPath, srv.DbPath)
	}

	file, err := ioutil.TempFile("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	srv = LoadConfig(file.Name())
	if srv == nil {
		t.Fatalf("LoadConfig should not return a nil value")
	}
	if srv.Address != shared.DefaultStorageAddress {
		t.Fatalf("srv.address should be '%s', instead is '%s'", shared.DefaultStorageAddress, srv.Address)
	}
	if srv.DbPath != shared.DefaultDbPath {
		t.Fatalf("srv.dbPath should be '%s', instead is '%s'", shared.DefaultDbPath, srv.DbPath)
	}

	srv.Address = "localhost:12345"
	srv.DbPath = "mydb"
	err = json.NewEncoder(file).Encode(srv)
	if err != nil {
		t.Fatal(err)
	}
	srv = LoadConfig(file.Name())
	if srv == nil {
		t.Fatalf("LoadConfig should not return a nil value")
	}
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
	srv := LoadConfig("")
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
