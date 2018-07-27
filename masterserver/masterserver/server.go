package masterserver

import (
	"encoding/json"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"

	"github.com/gorilla/mux"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/shared"
)

type Server struct {
	Address        string
	StorageAddress string
	quitCh         chan struct{}
	nodes          []shared.Node
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
		glog.Infof("Address is unset, using default address '%s'", shared.DefaultMasterserverAddress)
		srv.Address = shared.DefaultMasterserverAddress
	}
	if srv.StorageAddress == "" {
		glog.Infof("StorageAddress is unset, using default address '%s'", shared.DefaultStorageAddress)
		srv.StorageAddress = shared.DefaultStorageAddress
	}
	srv.quitCh = make(chan struct{})
}

func (srv *Server) Run() error {
	router := mux.NewRouter()
	// Create a route for our REST API on the method GET for list.
	router.HandleFunc("/list", srv.SendList).Methods("GET")
	listener, err := net.Listen("tcp", srv.Address)
	if err != nil {
		glog.Error(err)
		return err
	}
	defer listener.Close()
	go http.Serve(listener, router)
	<-srv.quitCh
	return nil
}

func (srv *Server) Stop() {
	close(srv.quitCh)
}

func (srv *Server) SendList(w http.ResponseWriter, r *http.Request) {
	nodes, err := srv.getList()
	if err != nil {
		// If for some reason we can't access the storage server, we'll send the list we have in cache
		// if we have one.
		if len(srv.nodes) > 0 {
			err = json.NewEncoder(w).Encode(srv.nodes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
	// We cache the result in case the storage server falls.
	srv.nodes = nodes
	err = json.NewEncoder(w).Encode(nodes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (srv *Server) getList() ([]shared.Node, error) {
	conn, err := net.Dial("tcp", srv.StorageAddress)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	rpcClt := rpc.NewClient(conn)
	defer rpcClt.Close()
	nodes := []shared.Node{}
	err = rpcClt.Call("Paths.ListFiles", &struct{}{}, &nodes)
	if err != nil {
		glog.Error(err)
		return []shared.Node{}, err
	}
	for i := range nodes {
		sort.Strings(nodes[i].Files)
	}
	return nodes, nil
}
