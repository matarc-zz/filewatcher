package main

type PathManager struct {
	events chan PathEvent
	list   []PathEvent
}

type PathEvent struct {
	Path  string
	Event Event
}

type Event uint32

const (
	Create Event = 1 << iota
	Remove
)

func (e Event) String() string {
	switch e {
	case Create:
		return "Create"
	case Remove:
		return "Remove"
	default:
		return "Unknown event"
	}
}

func NewPathManager(pathCh chan<- PathEvent) *PathManager {
	pm := new(PathManager)
	pm.events = make(chan PathEvent, 100)
	go pm.handleList(pathCh)
	return pm
}

func (pm *PathManager) GetEventsChan() chan<- PathEvent {
	return pm.events
}

func (pm *PathManager) handleList(pathCh chan<- PathEvent) {
	var buf *PathEvent
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
