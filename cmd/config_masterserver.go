// +build masterserver

package main

import (
	"github.com/matarc/filewatcher/masterserver"
	"github.com/matarc/filewatcher/shared"
)

const defaultCfgPath = "masterserver.conf"

func RunnableInstance() shared.Runnable {
	return new(masterserver.Server)
}
