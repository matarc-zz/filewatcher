package main

import "github.com/matarc/filewatcher/shared"

type PathManager struct {
	events chan shared.Operation
	list   []shared.Operation
}

func NewPathManager(pathCh chan<- []shared.Operation) *PathManager {
	pm := new(PathManager)
	pm.events = make(chan shared.Operation, 100)
	go pm.handleList(pathCh)
	return pm
}

func (pm *PathManager) GetEventsChan() chan<- shared.Operation {
	return pm.events
}

func (pm *PathManager) handleList(pathCh chan<- []shared.Operation) {
	var buf []shared.Operation
	dataSentCh := make(chan struct{})
	for {
		if len(buf) == 0 && len(pm.list) > 0 {
			buf = pm.list
			pm.list = []shared.Operation{}
			go func() {
				pathCh <- buf
				buf = []shared.Operation{}
				dataSentCh <- struct{}{}
			}()
		}
		select {
		case event := <-pm.events:
			pm.list = append(pm.list, event)
		case <-dataSentCh:
		}
	}
}
