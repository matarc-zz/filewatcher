package storage

import (
	"net"
	"net/rpc"
	"time"

	"github.com/boltdb/bolt"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Server struct {
	Address  string
	DbPath   string
	rpcSrv   *rpc.Server
	listener net.Listener
	db       *bolt.DB
}

func (srv *Server) Init() {
	if srv.Address == "" {
		glog.Infof("Address is unset, using default address '%s'", shared.DefaultStorageAddress)
		srv.Address = shared.DefaultStorageAddress
	}
	if srv.DbPath == "" {
		glog.Infof("DbPath is unset, using default path '%s'", shared.DefaultDbPath)
		srv.DbPath = shared.DefaultDbPath
	}
	srv.rpcSrv = rpc.NewServer()
}

func (srv *Server) Run() (err error) {
	srv.db, err = bolt.Open(srv.DbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		glog.Error(err)
		return
	}

	paths := new(shared.Paths)
	paths.Db = srv.db
	srv.rpcSrv.Register(paths)

	srv.listener, err = net.Listen("tcp", srv.Address)
	if err != nil {
		glog.Error(err)
		return
	}

	go func() {
		srv.rpcSrv.Accept(srv.listener)
	}()
	return nil
}

func (srv *Server) Stop() {
	srv.db.Close()
	srv.listener.Close()
}
