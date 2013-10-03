package eventhub

type EventFeed interface {
	Updates() <-chan *Event
	Close() error
}

//Queryable Data store
type DataBackend interface {
	Save(e *Event) error
	GetById(id int) (*Event, error)
	Query(q Query) ([]*Event, error)
}

type Broadcaster interface {
	Broadcast(e *Event)
	//Get Error chan
}

type DataService interface {
	SetBackend(d *DataBackend) error
	Run() error
}
