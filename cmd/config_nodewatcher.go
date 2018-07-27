// +build nodewatcher

package main

import (
	"github.com/matarc/filewatcher/nodewatcher/nodewatcher"
	"github.com/matarc/filewatcher/shared"
)

const defaultCfgPath = "nodewatcher.conf"

func RunnableInstance() shared.Runnable {
	return new(nodewatcher.Client)
}
