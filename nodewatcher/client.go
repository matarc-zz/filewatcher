package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

type Client struct {
	serverAddress string
	quitCh        chan struct{}
	buf           PathEvent
	bufIsEmpty    bool
}

var ErrQuit = fmt.Errorf("Quit")

func NewClient(serverAddress string) *Client {
	clt := new(Client)
	clt.serverAddress = serverAddress
	clt.quitCh = make(chan struct{})
	clt.bufIsEmpty = true
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

func (clt *Client) Run(pathCh <-chan PathEvent) {
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

func (clt *Client) sendList(conn net.Conn, pathCh <-chan PathEvent) {
	enc := gob.NewEncoder(conn)
	for {
		if clt.bufIsEmpty {
			select {
			case clt.buf = <-pathCh:
				clt.bufIsEmpty = false
			case <-clt.quitCh:
				return
			}
		}
		err := enc.Encode(clt.buf)
		if err != nil {
			glog.Error(err)
			break
		}
		clt.buf = PathEvent{}
		clt.bufIsEmpty = true
	}
}
