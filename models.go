package eventhub

import (
	"fmt"
	"time"
)

type Event struct {
	ID              int         `json:"id"`
	Key             string      `json:"key"`
	Created         time.Time   `json:"created"`
	Updated         time.Time   `json:"updated"`
	Payload         interface{} `json:"payload"`
	Description     string      `json:"description"`
	Importance      int         `json:"importance"`
	Origin          string      `json:"origin"`
	Entities        []string    `json:"entities"`
	OtherReferences []string    `json:"other_references"`
	Actors          []string    `json:"actors"`
	Tags            []string    `json:"tags"`
}

func (e Event) String() string {
	return fmt.Sprintf("%s, %v, %s, %v, %v", e.Key, e.Created, e.Origin, e.Entities, e.Actors)
}
