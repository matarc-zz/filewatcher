package shared

type Operation struct {
	Path  string
	Event Event
}

type Transaction struct {
	Id         string
	Operations []Operation
}

type Node struct {
	Id    string
	Files []string
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
