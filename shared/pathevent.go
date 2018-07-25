package shared

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
