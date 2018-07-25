package main

import "github.com/matarc/filewatcher/shared"

type PathManager struct {
	events chan shared.PathEvent
	list   []shared.PathEvent
}

func NewPathManager(pathCh chan<- shared.PathEvent) *PathManager {
	pm := new(PathManager)
	pm.events = make(chan shared.PathEvent, 100)
	go pm.handleList(pathCh)
	return pm
}

func (pm *PathManager) GetEventsChan() chan<- shared.PathEvent {
	return pm.events
}

func (pm *PathManager) handleList(pathCh chan<- shared.PathEvent) {
	var buf *shared.PathEvent
	dataSentCh := make(chan struct{})
	for {
		if buf == nil && len(pm.list) > 0 {
			buf = &pm.list[0]
			pm.list = pm.list[1:]
			go func() {
				pathCh <- *buf
				buf = nil
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
