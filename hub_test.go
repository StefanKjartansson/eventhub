package eventhub

import (
	"testing"
	"time"
)

type DummyFeed struct {
	events chan *Event
}

func (d DummyFeed) Updates() <-chan *Event {
	return d.events
}

func (d DummyFeed) Close() error {
	close(d.events)
	return nil
}

type FakeBroadCaster struct {
	events chan *Event
}

func (f FakeBroadCaster) Broadcast(e *Event) {
	f.events <- e
}

func TestHub(t *testing.T) {

	d := NewLocalMemoryStore()
	h := NewHub("Application", d)

	f1 := DummyFeed{make(chan *Event)}
	f2 := DummyFeed{make(chan *Event)}

	h.AddFeeds(f1, f2)

	b := FakeBroadCaster{make(chan *Event)}
	h.AddBroadcasters(b)

	count := 5
	ticker := time.NewTicker(1 * time.Millisecond)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				e := NewEvent(
					"myapp.user.login",
					nil,
					nil,
					"User foobar logged in",
					3,
					"myapp",
					[]string{"ns/foo", "ns/moo"},
					nil,
					nil,
					nil)

				t.Logf("Sending %+v to feed, count: %d", e, count)

				e.Description = "from feed 1"
				f1.events <- e
				e.Description = "from feed 2"
				f2.events <- e
				count--

				if count < 0 {
					close(quit)
				}
			case <-quit:
				t.Logf("Closing ticker")
				ticker.Stop()
				return
			}
		}
	}()

	go h.Run()

	for i := 0; i < 12; i++ {
		t.Logf("Broadcast: %+v", <-b.events)
	}
}
