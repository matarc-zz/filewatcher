package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/matarc/filewatcher/shared"

	"github.com/matarc/filewatcher/log"
)

var (
	config = flag.String("config", defaultCfgPath, "`Path` to configuration file")
)

func init() {
	flag.Parse()
}

func main() {
	var srv shared.Runnable

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)

Reboot:
	srv = RunnableInstance()
	shared.LoadConfig(*config, srv)
	err := srv.Run()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	for {
		select {
		case sig := <-sigCh:
			switch sig {
			case syscall.SIGUSR1:
				// Reload configuration
				log.Infof("Reloading configuration '%s'", *config)
				srv.Stop()
				goto Reboot
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGQUIT:
				// Terminate the program
				log.Info(sig)
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
