package straumur

type EventFeed interface {
	Updates() <-chan *Event
	Close() error
}

//Queryable Data store
type DataBackend interface {
	Save(e *Event) error
	GetById(id int) (*Event, error)
	Query(q Query) ([]*Event, error)
	AggregateType(q Query, s string) (map[string]int, error)
}

type Broadcaster interface {
	Broadcast(e *Event)
	Run(ec chan error)
}

type DataService interface {
	Run(d DataBackend, ec chan error)
}
