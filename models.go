package eventhub

import (
	"fmt"
	"time"
)

type Event struct {
	ID              int
	Key             string
	Created         time.Time
	Updated         time.Time
	Payload         interface{}
	Description     string
	Importance      int
	Origin          string
	Entities        []string
	OtherReferences []string
	Actors          []string
	Tags            []string
}

func (e Event) String() string {
	return fmt.Sprintf("%s, %v, %s, %v, %v", e.Key, e.Created, e.Origin, e.Entities, e.Actors)
}
