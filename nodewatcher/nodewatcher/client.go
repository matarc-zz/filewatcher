package nodewatcher

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Client struct {
	StorageAddress string
	quitCh         chan struct{}
	buf            []shared.Operation
	Id             string
	Dir            string
	watcher        *Watcher
	pm             *PathManager
}

func (clt *Client) Init() {
	if clt.Id == "" {
		id, err := os.Hostname()
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infof("Id is unset, using hostname '%s'", id)
			clt.Id = id
		}
	}
	if clt.Dir == "" {
		clt.Dir = os.TempDir()
		glog.Infof("Dir is unset, using tempdir '%s'", clt.Dir)
	}
	if clt.StorageAddress == "" {
		glog.Infof("StorageAddress is unset, using default address '%s'", shared.DefaultStorageAddress)
		clt.StorageAddress = shared.DefaultStorageAddress
	}
	clt.quitCh = make(chan struct{})
}

func (clt *Client) Stop() {
	close(clt.quitCh)
	clt.watcher.Stop()
	clt.pm.Stop()
}

func (clt *Client) dial() (conn net.Conn, err error) {
	for {
		conn, err = net.Dial("tcp", clt.StorageAddress)
		if err == nil {
			return
		}
		glog.Error(err)
		// We wait 10 seconds before attempting a connection again in order to not use 100% of the CPU
		// if the server is down.
		select {
		case <-clt.quitCh:
			return nil, shared.ErrQuit
		case <-time.After(time.Second * 10):
		}
	}
	return
}

func (clt *Client) Run() error {
	pathCh := make(chan []shared.Operation)
	clt.pm = NewPathManager(pathCh)
	clt.watcher = NewWatcher(clt.Dir)
	if clt.watcher == nil {
		return fmt.Errorf("Watcher couldn't be initialized")
	}
	go func() {
		clt.watcher.WatchDir(pathCh)
		clt.watcher.HandleFileEvents(pathCh)
	}()
	go clt.run(pathCh)
	return nil
}

func (clt *Client) run(pathCh <-chan []shared.Operation) {
	for {
		select {
		case <-clt.quitCh:
			return
		default:
		}
		func() {
			conn, err := clt.dial()
			if err == shared.ErrQuit {
				return
			}
			defer conn.Close()
			clt.sendList(conn, pathCh)
		}()
	}
}

func (clt *Client) sendList(conn net.Conn, pathCh <-chan []shared.Operation) {
	rpcClt := rpc.NewClient(conn)
	defer rpcClt.Close()
	for {
		if len(clt.buf) == 0 {
			select {
			case clt.buf = <-pathCh:
			case <-clt.quitCh:
				return
			}
		}
		transaction := &shared.Transaction{Id: clt.Id, Operations: clt.buf}
		reply := new(shared.Transaction)
		err := rpcClt.Call("Paths.Update", transaction, reply)
		if err != nil {
			glog.Error(err)
			break
		}
		clt.buf = []shared.Operation{}
	}
}
