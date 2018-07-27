package masterserver

import (
	"encoding/json"
	"net"
	"net/http"
	"net/rpc"
	"sort"

	"github.com/gorilla/mux"

	"github.com/matarc/filewatcher/log"
	"github.com/matarc/filewatcher/shared"
)

type Server struct {
	Address        string
	StorageAddress string
	nodes          []shared.Node
	listener       net.Listener
}

func (srv *Server) Init() {
	if srv.Address == "" {
		log.Infof("Address is unset, using default address '%s'", shared.DefaultMasterserverAddress)
		srv.Address = shared.DefaultMasterserverAddress
	}
	if srv.StorageAddress == "" {
		log.Infof("StorageAddress is unset, using default address '%s'", shared.DefaultStorageAddress)
		srv.StorageAddress = shared.DefaultStorageAddress
	}
}

func (srv *Server) Run() (err error) {
	router := mux.NewRouter()
	// Create a route for our REST API on the method GET for list.
	router.HandleFunc("/list", srv.SendList).Methods("GET")
	srv.listener, err = net.Listen("tcp", srv.Address)
	if err != nil {
		log.Error(err)
		return err
	}
	go http.Serve(srv.listener, router)
	return nil
}

func (srv *Server) Stop() {
	srv.listener.Close()
}

func (srv *Server) SendList(w http.ResponseWriter, r *http.Request) {
	nodes, err := srv.getList()
	if err != nil {
		// If for some reason we can't access the storage server, we'll send the list we have in cache
		// if we have one.
		log.Error(err)
		if len(srv.nodes) > 0 {
			err = json.NewEncoder(w).Encode(srv.nodes)
			if err != nil {
				log.Error(err)
				http.Error(w, "Encoding issue", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Server unreachable", http.StatusBadGateway)
			return
		}
	}
	// We cache the result in case the storage server falls.
	srv.nodes = nodes
	err = json.NewEncoder(w).Encode(nodes)
	if err != nil {
		log.Error(err)
		http.Error(w, "Encoding issue", http.StatusInternalServerError)
		return
	}
}

func (srv *Server) getList() ([]shared.Node, error) {
	conn, err := net.Dial("tcp", srv.StorageAddress)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	rpcClt := rpc.NewClient(conn)
	defer rpcClt.Close()
	nodes := []shared.Node{}
	err = rpcClt.Call("Paths.ListFiles", &struct{}{}, &nodes)
	if err != nil {
		log.Error(err)
		return []shared.Node{}, err
	}
	for i := range nodes {
		sort.Strings(nodes[i].Files)
	}
	return nodes, nil
}
