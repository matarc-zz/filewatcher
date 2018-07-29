package storage

import (
	"net"
	"net/rpc"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"

	"github.com/matarc/filewatcher/log"
	"github.com/matarc/filewatcher/shared"
)

type Server struct {
	Address  string
	DbPath   string
	rpcSrv   *rpc.Server
	listener net.Listener
	db       *bolt.DB
}

// Init initialises the server.
func (srv *Server) Init() {
	if srv.Address == "" {
		log.Infof("Address is unset, using default address '%s'", shared.DefaultStorageAddress)
		srv.Address = shared.DefaultStorageAddress
	}
	if srv.DbPath == "" {
		log.Infof("DbPath is unset, using default path '%s'", shared.DefaultDbPath)
		srv.DbPath = shared.DefaultDbPath
	}
	srv.rpcSrv = rpc.NewServer()
	srv.DbPath = filepath.Clean(srv.DbPath)
}

// Run starts the storage server by opening its database `db` and listening on `Address` to
// start the RPC.Server.
// It returns an error if it fails to do any of those two actions.
func (srv *Server) Run() (err error) {
	log.Infof("Opening database '%s'", srv.DbPath)
	srv.db, err = bolt.Open(srv.DbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Error(err)
		return
	}

	paths := new(shared.Paths)
	paths.Db = srv.db
	srv.rpcSrv.Register(paths)

	log.Infof("Listening on '%s'", srv.Address)
	srv.listener, err = net.Listen("tcp", srv.Address)
	if err != nil {
		log.Error(err)
		return
	}

	go func() {
		srv.rpcSrv.Accept(srv.listener)
	}()
	return nil
}

// Stop stops the server.
func (srv *Server) Stop() {
	if srv.db != nil {
		srv.db.Close()
	}
	if srv.listener != nil {
		srv.listener.Close()
	}
}
