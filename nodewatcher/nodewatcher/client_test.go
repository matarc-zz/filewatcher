package nodewatcher

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/matarc/filewatcher/shared"
)

func TestRun(t *testing.T) {
	// Setup
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	myDir := filepath.Join(rootDir, "my")
	err = os.Mkdir(myDir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(myDir, "path"))
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	dbPath := filepath.Join(rootDir, "mydb")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clt := new(Client)
	shared.LoadConfig("", clt)
	clt.Id = "client1"
	clt.Dir = myDir
	listener, err := net.Listen("tcp", shared.DefaultStorageAddress)
	if err != nil {
		t.Fatal(err)
	}
	go clt.Run()
	rpcSrv := rpc.NewServer()
	paths := new(shared.Paths)
	paths.Db = db
	rpcSrv.Register(paths)
	go rpcSrv.Accept(listener)

	// Test
	// Ugly but wait for the database to write changes
	time.Sleep(time.Second)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("client1"))
		if b == nil {
			return fmt.Errorf("Bucket '%s' doesn't exist", "client1")
		}
		val := b.Get([]byte("my/path"))
		if val == nil {
			return fmt.Errorf("'my/path' should exist in database")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInit(t *testing.T) {
	clt := new(Client)
	clt.Init()
	if clt.quitCh == nil {
		t.Fatalf("quitCh should not be a nil channel")
	}
	if clt.StorageAddress != shared.DefaultStorageAddress {
		t.Fatalf("StorageAddress should be '%s', instead is '%s'", shared.DefaultStorageAddress, clt.StorageAddress)
	}
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	if clt.Id != hostname {
		t.Fatalf("Id should be '%s', instead is '%s'", hostname, clt.Id)
	}
	tmpDir := os.TempDir()
	if clt.Dir != tmpDir {
		t.Fatalf("Dir should be '%s', instead is '%s'", tmpDir, clt.Id)
	}

	clt = new(Client)
	clt.Id = "1"
	clt.Dir = "/my"
	clt.StorageAddress = "localhost:12345"
	clt.Init()
	if clt.Id != "1" {
		t.Fatalf("Id should be '1', instead is '%s", clt.Id)
	}
	if clt.Dir != "/my" {
		t.Fatalf("Dir should be '/my', instead is '%s", clt.Dir)
	}
	if clt.StorageAddress != "localhost:12345" {
		t.Fatalf("StorageAddress should be 'localhost:12345', instead is '%s", clt.StorageAddress)
	}
}
