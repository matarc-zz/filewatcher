package storage

import (
	"encoding/json"
	"net"
	"net/rpc"
	"os"
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

func LoadConfig(cfgPath string) *Server {
	srv := new(Server)
	file, err := os.Open(cfgPath)
	if err != nil {
		glog.Errorf("Can't open '%s', using default configuration instead", cfgPath)
	} else {
		err = json.NewDecoder(file).Decode(srv)
		if err != nil {
			glog.Errorf("Can't decode '%s' as a json file, using default configuration instead", cfgPath)
		}
	}
	srv.init()
	return srv
}

func (srv *Server) init() {
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
