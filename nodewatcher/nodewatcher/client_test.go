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
	dbPath := filepath.Join(rootDir, "mydb")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clt := NewClient("localhost:12345", "client1")
	pathCh := make(chan []shared.Operation)
	listener, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		t.Fatal(err)
	}
	go clt.Run(pathCh)
	rpcSrv := rpc.NewServer()
	paths := new(shared.Paths)
	paths.Db = db
	rpcSrv.Register(paths)
	go rpcSrv.Accept(listener)

	// Test
	pathCh <- []shared.Operation{shared.Operation{"/my/path", shared.Create}}
	// Ugly but wait for the database to write changes
	time.Sleep(time.Second)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("client1"))
		val := b.Get([]byte("/my/path"))
		if val == nil {
			return fmt.Errorf("'/my/path' should exist in database")
		}
		return nil
	})
}

func TestNewClient(t *testing.T) {
	clt := NewClient("localhost:12345", "client1")
	if clt == nil {
		t.Fatalf("NewClient should not return a nil value")
	}
	if len(clt.buf) != 0 {
		t.Fatalf("buf should be empty")
	}
	if clt.quitCh == nil {
		t.Fatalf("quitCh should not be a nil channel")
	}
	if clt.StorageAddress != "localhost:12345" {
		t.Fatalf("serverAddress should be 'localhost:12345', instead is '%s'", clt.StorageAddress)
	}
	if clt.Id != "client1" {
		t.Fatalf("id should be 'client1', instead is '%s'", clt.Id)
	}
}
