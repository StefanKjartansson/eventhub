package eventhub

import (
	"fmt"
	"time"
)

type Event struct {
	ID              int         `json:"id"`
	Key             string      `json:"key"`
	KeyParams       interface{} `json:"key_params"`
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

func NewEvent(key string, keyParams interface{}, payload interface{}, description string, importance int,
	origin string, entities []string, otherReferences []string, actors []string, tags []string) *Event {

	return &Event{
		Key:             key,
		KeyParams:       keyParams,
		Created:         time.Now(),
		Description:     description,
		Importance:      importance,
		Origin:          origin,
		Entities:        entities,
		OtherReferences: otherReferences,
		Actors:          actors,
		Tags:            tags,
	}
}

func (e Event) String() string {
	return fmt.Sprintf("%s, %v, %s, %v, %v", e.Key, e.Created, e.Origin, e.Entities, e.Actors)
}
