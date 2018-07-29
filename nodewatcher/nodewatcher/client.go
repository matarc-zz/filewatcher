package nodewatcher

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"

	"github.com/matarc/filewatcher/log"
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

// Init initialize the client.
func (clt *Client) Init() {
	if clt.Id == "" {
		id, err := os.Hostname()
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("Id is unset, using hostname '%s'", id)
			clt.Id = id
		}
	}
	if clt.Dir == "" {
		clt.Dir = os.TempDir()
		log.Infof("Dir is unset, using tempdir '%s'", clt.Dir)
	}
	if clt.StorageAddress == "" {
		log.Infof("StorageAddress is unset, using default address '%s'", shared.DefaultStorageAddress)
		clt.StorageAddress = shared.DefaultStorageAddress
	}
	clt.quitCh = make(chan struct{})
	clt.Dir = filepath.Clean(clt.Dir)
}

// Stop stops the client, it no longer monitor the directory after that.
func (clt *Client) Stop() {
	if clt.quitCh != nil {
		close(clt.quitCh)
		clt.quitCh = nil
	}
	if clt.watcher != nil {
		clt.watcher.Stop()
	}
	clt.pm.Stop()
}

// dial attempts to connect to the storage server, it is a blocking until it gets a connection,
// or until `Stop` is called.
// If it fails to establish a connection, it will try again after 10 seconds.
// Return a connection upon establishing one or `shared.ErrQuit` if the client was stopped.
func (clt *Client) dial() (conn net.Conn, err error) {
	log.Info("dial")
	for {
		conn, err = net.Dial("tcp", clt.StorageAddress)
		if err == nil {
			return
		}
		log.Error(err)
		// We wait 10 seconds before attempting a connection again in order to not use 100% of the CPU
		// if the server is down.
		log.Info("Waiting 10 seconds")
		select {
		case <-clt.quitCh:
			return nil, shared.ErrQuit
		case <-time.After(time.Second * 10):
		}
	}
	return
}

// Run starts the client, walking through the directory and its subdirectories to list all files.
// It will delete the remote list on the storage server before sending the current list.
// After sending the list it will sends all updates on any file within the directory or its subdirectories.
// It returns an error if the path in `Dir` is not a directory or if it's not watchable.
func (clt *Client) Run() error {
	pathCh := make(chan []shared.Operation)
	clt.pm = NewPathManager(pathCh)
	clt.watcher = NewWatcher(clt.Dir)
	if clt.watcher == nil {
		return fmt.Errorf("Watcher couldn't be initialized")
	}
	err := clt.watcher.CheckDir()
	if err != nil {
		return err
	}
	go func() {
		clt.watcher.WatchDir(pathCh)
		clt.watcher.HandleFileEvents(pathCh)
	}()
	go clt.run(pathCh)
	return nil
}

// run tries to connect to the storage server.
// It will delete the remote list on the storage server upon the first connection,
// and then just keep establishing a new one whenever the connection is broken until `Stop` is called.
func (clt *Client) run(pathCh <-chan []shared.Operation) {
	for {
		select {
		case <-clt.quitCh:
			return
		default:
		}
		conn, err := clt.dial()
		if err == shared.ErrQuit {
			return
		}
		err = clt.deleteRemoteList(conn)
		conn.Close()
		// Ugly but until this issue is fixed, not much we can do about it
		// https://github.com/golang/go/issues/23340
		if err == nil || err.Error() == bolt.ErrBucketNotFound.Error() {
			break
		}
		log.Error(err)
	}

	for {
		select {
		case <-clt.quitCh:
			return
		default:
		}
		conn, err := clt.dial()
		if err == shared.ErrQuit {
			return
		}
		clt.sendList(conn, pathCh)
		conn.Close()
	}
}

// deleteRemoteList makes an RPC to delete the list on the storage server.
// It returns an error if it fails to do so.
func (clt *Client) deleteRemoteList(conn net.Conn) error {
	rpcClt := rpc.NewClient(conn)
	defer rpcClt.Close()
	return rpcClt.Call("Paths.DeleteList", clt.Id, &struct{}{})
}

// sendList makes an RPC to send the list on the storage server as well as all,
// updates within the watched directory and its subdirectories.
func (clt *Client) sendList(conn net.Conn, pathCh <-chan []shared.Operation) {
	log.Info("sendList")
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
		log.Infof("Operations : '%v'", clt.buf)
		err := rpcClt.Call("Paths.Update", transaction, reply)
		if err != nil {
			log.Error(err)
			break
		}
		clt.buf = []shared.Operation{}
	}
}
