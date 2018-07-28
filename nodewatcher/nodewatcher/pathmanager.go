package nodewatcher

import "github.com/matarc/filewatcher/shared"

type PathManager struct {
	operations chan []shared.Operation
	list       []shared.Operation
	quitCh     chan struct{}
}

func NewPathManager(pathCh chan<- []shared.Operation) *PathManager {
	pm := new(PathManager)
	pm.operations = make(chan []shared.Operation, 10)
	pm.quitCh = make(chan struct{})
	go pm.handleList(pathCh)
	return pm
}

func (pm *PathManager) GetChan() chan<- []shared.Operation {
	return pm.operations
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
		case operation := <-pm.operations:
			pm.list = append(pm.list, operation...)
		case <-dataSentCh:
		case <-pm.quitCh:
			return
		}
	}
}

func (pm *PathManager) Stop() {
	close(pm.quitCh)
}
