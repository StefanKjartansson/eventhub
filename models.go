package eventhub

import (
	"time"
)

type Event struct {
	ID              int
	Key             string
	Created         time.Time
	Payload         interface{}
	Description     string
	Importance      int
	Origin          string
	Entities        []string
	OtherReferences []string
	Actors          []string
	Tags            []string
}
