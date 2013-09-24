package eventhub

type DataBackend interface {
	Save(e *Event) error
	GetById(id int) (*Event, error)
	FilterBy(m map[string]interface{}) ([]*Event, error)
}

type EventFeed interface {
	Updates() <-chan Event
	Close() error
}

type Broadcaster interface {
	Register(client int)
	Constrict(client int, parameter, value string) //only broadcast certain events
	Listen() error
	Stop() error
}
