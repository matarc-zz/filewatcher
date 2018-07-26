package main

import (
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Client struct {
	serverAddress string
	quitCh        chan struct{}
	buf           []shared.Operation
	id            string
}

var ErrQuit = fmt.Errorf("Quit")

func NewClient(serverAddress, id string) *Client {
	clt := new(Client)
	clt.serverAddress = serverAddress
	clt.quitCh = make(chan struct{})
	clt.id = id
	return clt
}

func (clt *Client) Stop() {
	close(clt.quitCh)
}

func (clt *Client) dial(network, address string) (conn net.Conn, err error) {
	for {
		conn, err = net.Dial("tcp", clt.serverAddress)
		if err == nil {
			return
		}
		glog.Error(err)
		// We wait 10 seconds before attempting a connection again in order to not use 100% of the CPU
		// if the server is down.
		select {
		case <-clt.quitCh:
			return nil, ErrQuit
		case <-time.After(time.Second * 10):
		}
	}
	return
}

func (clt *Client) Run(pathCh <-chan []shared.Operation) {
	for {
		select {
		case <-clt.quitCh:
			return
		default:
		}
		func() {
			conn, err := clt.dial("tcp", clt.serverAddress)
			if err == ErrQuit {
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
		transaction := &shared.Transaction{Id: clt.id, Operations: clt.buf}
		reply := new(shared.Transaction)
		err := rpcClt.Call("Paths.Update", transaction, reply)
		if err != nil {
			glog.Error(err)
			break
		}
		clt.buf = []shared.Operation{}
	}
}
