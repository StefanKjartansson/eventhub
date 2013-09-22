package eventhub

type DataBackend interface {
	Save(e *Event) error
	GetById(id int) (*Event, error)
	FilterBy(m map[string]interface{}) ([]*Event, error)
}

type EventFeed interface {
	Updates() <-chan Event
	Close() error
	loop() error
}
