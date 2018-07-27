package shared

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

type server struct {
	Address        string
	StorageAddress string
}

func (s *server) Init()      {}
func (s *server) Run() error { return nil }
func (s *server) Stop()      {}

func TestLoadConfig(t *testing.T) {
	file, err := ioutil.TempFile("", "filewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	srv := new(server)
	LoadConfig(file.Name(), srv)
	if srv == nil {
		t.Fatalf("LoadConfig should not return nil")
	}
	if srv.Address != "" {
		t.Fatalf("address should be '', instead is '%s'", srv.Address)
	}
	if srv.StorageAddress != "" {
		t.Fatalf("storageAddress should be '', instead is '%s'", srv.StorageAddress)
	}

	srv.Address = "localhost:12345"
	srv.StorageAddress = "localhost:54321"
	err = json.NewEncoder(file).Encode(srv)
	if err != nil {
		t.Fatal(err)
	}
	srv = new(server)
	LoadConfig(file.Name(), srv)
	if srv == nil {
		t.Fatalf("LoadConfig should not return nil")
	}
	if srv.Address != "localhost:12345" {
		t.Fatalf("address should be 'localhost:12345', instead is '%s'", srv.Address)
	}
	if srv.StorageAddress != "localhost:54321" {
		t.Fatalf("storageAddress should be 'localhost:54321', instead is '%s'", srv.StorageAddress)
	}
}
