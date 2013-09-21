package eventhub

type DataBackend interface {
	Save(e *Event) error
	GetById(id int) (*Event, error)
	//FilterByKey(key string) (*[]Event, error)
}

