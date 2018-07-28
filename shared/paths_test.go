package shared

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/matarc/filewatcher/log"
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
		log.Error(err)
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

func TestListFiles(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	dbPath := filepath.Join(rootDir, "mydb")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Error(err)
		return
	}
	defer db.Close()
	err = db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("1"))
		if err != nil {
			return err
		}
		err = b.Put([]byte("/my/path1"), []byte{})
		if err != nil {
			return err
		}
		err = b.Put([]byte("/my/path2"), []byte{})
		if err != nil {
			return err
		}
		b, err = tx.CreateBucket([]byte("2"))
		if err != nil {
			return err
		}
		err = b.Put([]byte("/your/path1"), []byte{})
		if err != nil {
			return err
		}
		err = b.Put([]byte("/your/path2"), []byte{})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	paths := new(Paths)
	paths.Db = db
	list := []Node{}
	err = paths.ListFiles(&struct{}{}, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("list should have '2' nodes, instead has '%d'", len(list))
	}
	if list[0].Id != "1" {
		t.Fatalf("First node should have id '1', instead has '%s'", list[0].Id)
	}
	if len(list[0].Files) != 2 {
		t.Fatalf("First node should have '2' files, instead has '%d'", len(list[0].Files))
	}
	m := make(map[string]bool)
	m["/my/path1"] = true
	m["/my/path2"] = true
	for _, path := range list[0].Files {
		if !m[path] {
			t.Fatalf("'%s' shouldn't be in the database", path)
		}
		delete(m, path)
	}
	if len(m) != 0 {
		t.Fatalf("Some path were not inserted in the database '%v'", m)
	}

	if list[1].Id != "2" {
		t.Fatalf("First node should have id '2', instead has '%s'", list[1].Id)
	}
	if len(list[1].Files) != 2 {
		t.Fatalf("First node should have '2' files, instead has '%d'", len(list[1].Files))
	}
	m["/your/path1"] = true
	m["/your/path2"] = true
	for _, path := range list[1].Files {
		if !m[path] {
			t.Fatalf("'%s' shouldn't be in the database", path)
		}
		delete(m, path)
	}
	if len(m) != 0 {
		t.Fatalf("Some path were not inserted in the database '%v'", m)
	}
}

func TestDeleteList(t *testing.T) {
	rootDir, err := ioutil.TempDir("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)
	dbPath := filepath.Join(rootDir, "mydb")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Error(err)
		return
	}
	defer db.Close()

	paths := new(Paths)
	paths.Db = db
	err = paths.DeleteList("1", &struct{}{})
	if err != bolt.ErrBucketNotFound {
		t.Fatal(err)
	}

	err = db.Batch(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("1"))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	err = paths.DeleteList("1", &struct{}{})
	if err != nil {
		t.Fatal(err)
	}
}
