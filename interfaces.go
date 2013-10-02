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
	Query(q Query) ([]*Event, error)
}

type Broadcaster interface {
	Register(client int)
	Constrict(client int, parameter, value string) //only broadcast certain events
	Listen() error
	Stop() error
}
