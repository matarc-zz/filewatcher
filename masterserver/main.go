package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/matarc/filewatcher/masterserver/masterserver"
)

var (
	config = flag.String("config", "masterserver.conf", "`Path` to configuration file")
)

func init() {
	flag.Parse()
}

func main() {
	var srv *masterserver.Server

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)

Reboot:
	srv = masterserver.LoadConfig(*config)
	err := srv.Run()
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}
	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGUSR1:
				// Reload configuration
				glog.Infof("Reloading configuration '%s'", *config)
				srv.Stop()
				goto Reboot
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGQUIT:
				// Terminate the program
				glog.Info(sig)
				srv.Stop()
				code := 0
				sigCode, ok := sig.(syscall.Signal)
				if ok {
					code = int(sigCode)
				}
				os.Exit(128 + code)
			}
		}
	}
}
