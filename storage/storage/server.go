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
	Address string
	quitCh  chan struct{}
	DbPath  string
	rpcSrv  *rpc.Server
}

func NewServer(address, dbPath string) *Server {
	srv := new(Server)
	srv.rpcSrv = rpc.NewServer()
	srv.Address = address
	srv.quitCh = make(chan struct{})
	srv.DbPath = dbPath
	return srv
}

func (srv *Server) Run() {
	db, err := bolt.Open(srv.DbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	paths := new(shared.Paths)
	paths.Db = db
	srv.rpcSrv.Register(paths)

	listener, err := net.Listen("tcp", srv.Address)
	if err != nil {
		glog.Error(err)
		return
	}
	defer listener.Close()

	go func() {
		srv.rpcSrv.Accept(listener)
	}()

	<-srv.quitCh
}

func (srv *Server) Stop() {
	close(srv.quitCh)
}
