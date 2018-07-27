package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"

	"github.com/matarc/filewatcher/log"
	"github.com/matarc/filewatcher/masterserver/masterserver"
	"github.com/matarc/filewatcher/nodewatcher/nodewatcher"
	"github.com/matarc/filewatcher/storage/storage"
)

var (
	config         = flag.String("config", "", "Create configuration for `[masterserver|nodewatcher|storage]`")
	address        = flag.String("address", "", "Address in the form `host:port` on which to listen (masterserver and storage only)")
	storageAddress = flag.String("storageAddress", "", "Address in the form `host:port` to dial to query the storage server (masterserver and nodewatcher only)")
	id             = flag.String("id", "", "Id for the client (nodewatcher only)")
	dbpath         = flag.String("dbpath", "mydb.bolt", "`Path` to the database (storage only)")
	dir            = flag.String("dir", "", "`Path` to the directory that must be watched (nodewatcher only)")
)

func init() {
	flag.Parse()
}

func main() {
	var buf []byte
	var err error
	switch *config {
	case "masterserver":
		cfg := masterserver.Server{Address: *address, StorageAddress: *storageAddress}
		buf, err = json.Marshal(cfg)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

	case "nodewatcher":
		cfg := nodewatcher.Client{StorageAddress: *storageAddress, Id: *id, Dir: *dir}
		buf, err = json.Marshal(cfg)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	case "storage":
		cfg := storage.Server{Address: *address, DbPath: *dbpath}
		buf, err = json.Marshal(cfg)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
	var out bytes.Buffer
	json.Indent(&out, buf, "", "\t")
	out.WriteTo(os.Stdout)
}
