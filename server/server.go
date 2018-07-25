package main

import (
	"encoding/gob"
	"net"
	"time"

	"github.com/boltdb/bolt"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Server struct {
	address string
	quitCh  chan struct{}
	dbPath  string
	db      *bolt.DB
}

func NewServer(address, dbPath string) *Server {
	srv := new(Server)
	srv.address = address
	srv.quitCh = make(chan struct{})
	srv.dbPath = dbPath
	return srv
}

func (srv *Server) Run() {
	var err error
	srv.db, err = bolt.Open(srv.dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		glog.Error(err)
		return
	}
	defer srv.db.Close()
	listener, err := net.Listen("tcp", srv.address)
	if err != nil {
		glog.Error(err)
		return
	}
	defer listener.Close()
	for {
		select {
		case <-srv.quitCh:
			return
		default:
		}
		conn, err := listener.Accept()
		if err != nil {
			glog.Error(err)
			continue
		}
		go srv.handleConnection(conn)
	}
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	dec := gob.NewDecoder(conn)
	var packet shared.Packet
	for {
		select {
		case <-srv.quitCh:
			return
		default:
		}
		err := dec.Decode(&packet)
		if err != nil {
			glog.Error(err)
			return
		}
		srv.updateDb(packet)
		if err != nil {
			glog.Error(err)
			return
		}
	}
}

func (srv *Server) updateDb(packet shared.Packet) error {
	return srv.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(packet.Id))
		if err != nil {
			return err
		}
		if packet.PathEvent.Event&shared.Create == shared.Create {
			err = b.Put([]byte(packet.PathEvent.Path), []byte{})
			if err != nil {
				return err
			}
		} else if packet.PathEvent.Event&shared.Remove == shared.Create {
			err = b.Delete([]byte(packet.PathEvent.Path))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
