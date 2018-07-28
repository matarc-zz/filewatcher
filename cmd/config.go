// +build !storage
// +build !nodewatcher
// +build !masterserver

package main

import (
	"github.com/matarc/filewatcher/shared"
)

const defaultCfgPath = "decoy.conf"

type server struct{}

func (s *server) Init()      {}
func (s *server) Run() error { return nil }
func (s *server) Stop()      {}

func RunnableInstance() shared.Runnable {
	return new(server)
}
