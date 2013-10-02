package eventhub

import (
	"testing"
)

type DummyFeed struct {
	events chan *Event
}

func (d DummyFeed) Updates() <-chan *Event {
	return d.events
}

func (d DummyFeed) Close() error {
	return nil
}

type FakeBroadCaster struct {
}

func (d FakeBroadCaster) Broadcast(e *Event) error {
	return nil
}

func TestHub(t *testing.T) {

	d := NewLocalMemoryStore()
	h := NewHub("Application", d)

	df1 := DummyFeed{make(chan *Event)}
	df2 := DummyFeed{make(chan *Event)}

	//Adds feeds to []EventFeeds
	h.AddFeed(df1)
	h.AddFeed(df2)

	b1 := FakeBroadCaster{}

	h.AddBroadcaster(b1)

	h.Run()
}
