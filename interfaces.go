package eventhub

type EventFeed interface {
	Updates() <-chan *Event
	Close() error
}

//Queryable Data store
type DataBackend interface {
	EventFeed
	Save(e *Event) error
	GetById(id int) (*Event, error)
	FilterBy(m map[string]interface{}) ([]*Event, error)
}

type Broadcaster interface {
	Broadcast(e *Event) error
}

type DataService interface {
	SetBackend(d *DataBackend) error
	Run() error
}
