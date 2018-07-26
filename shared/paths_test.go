package shared

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

func TestUpdate(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	dbPath := filepath.Join(rootDir, "mydb")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	paths := new(Paths)
	paths.Db = db
	reply := new(Transaction)
	transactions := new(Transaction)

	transactions.Id = "1"
	transactions.Operations = append(transactions.Operations, Operation{Path: "/my/path", Event: Create})
	err = paths.Update(transactions, reply)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Operations) != 1 {
		t.Fatalf("Reply should only have '1' operation, instead has '%d'", len(reply.Operations))
	}
	if reply.Operations[0].Path != "/my/path" {
		t.Fatalf("Path should be '/my/path', instead is '%s'", reply.Operations[0].Path)
	}
	if reply.Operations[0].Event&Create != Create {
		t.Fatalf("Event should be 'Create', instead is '%s'", reply.Operations[0].Event)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("1"))
		val := b.Get([]byte("/my/path"))
		if val == nil {
			return fmt.Errorf("'/my/path' should be in database")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	reply = new(Transaction)
	transactions.Operations[0].Event = Remove
	transactions.Operations = append(transactions.Operations, Operation{Path: "/your/path", Event: Create})
	err = paths.Update(transactions, reply)
	if err != nil {
		t.Fatal(err)
	}
	if len(reply.Operations) != 2 {
		t.Fatalf("Reply should only have '2' operation, instead has '%d'", len(reply.Operations))
	}
	if reply.Operations[0].Path != "/my/path" {
		t.Fatalf("Path should be '/my/path', instead is '%s'", reply.Operations[0].Path)
	}
	if reply.Operations[0].Event&Remove != Remove {
		t.Fatalf("Event should be 'Remove', instead is '%s'", reply.Operations[0].Event)
	}
	if reply.Operations[1].Path != "/your/path" {
		t.Fatalf("Path should be '/your/path', instead is '%s'", reply.Operations[0].Path)
	}
	if reply.Operations[1].Event&Create != Create {
		t.Fatalf("Event should be 'Create', instead is '%s'", reply.Operations[0].Event)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("1"))
		val := b.Get([]byte("/my/path"))
		if val != nil {
			return fmt.Errorf("'/my/path' should not be in database")
		}
		val = b.Get([]byte("/your/path"))
		if val == nil {
			return fmt.Errorf("'/your/path' should be in database")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
