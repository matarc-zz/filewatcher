// +build storage

package main

import (
	"github.com/matarc/filewatcher/shared"
	"github.com/matarc/filewatcher/storage"
)

const defaultCfgPath = "storage.conf"

func RunnableInstance() shared.Runnable {
	return new(storage.Server)
}
