package straumur

import (
	"errors"
	"fmt"
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

func (f FakeBroadCaster) Run(er chan error) {
}

type FakeDataService struct {
	d DataBackend
}

func (f FakeDataService) Run(d DataBackend, ec chan error) {
	f.d = d
	if f.d == nil {
		ec <- fmt.Errorf("FakeDataService started with nil DataBackend")
	}
}

func Ticker(count int, efs ...DummyFeed) <-chan struct{} {
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

				for idx, f := range efs {
					e.Description = fmt.Sprintf("from feed %d", idx+1)
					f.events <- e
				}
				count--
				if count < 0 {
					close(quit)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit
}

func TestHub(t *testing.T) {

	t.Skip()
	d := NewLocalMemoryStore()
	h := NewHub(d)

	f1 := DummyFeed{make(chan *Event)}
	f2 := DummyFeed{make(chan *Event)}

	h.AddFeeds(f1, f2)

	b := FakeBroadCaster{make(chan *Event)}
	h.AddBroadcasters(b)

	fds := FakeDataService{}
	h.AddDataServices(fds)

	wait := Ticker(5, f1, f2)

	go h.Run()

	<-wait

	for i := 0; i < 12; i++ {
		t.Logf("Broadcast: %+v", <-b.events)
	}

}

func TestErrClose(t *testing.T) {

	ErrSome := errors.New("some error")
	d := NewLocalMemoryStore()
	h := NewHub(d)
	f1 := DummyFeed{make(chan *Event)}
	h.AddFeeds(f1)

	h.RegisterProcessor("pre", "myapp*", func(e *Event) error {
		return ErrSome
	})

	wait := Ticker(1, f1)

	go h.Run()

	<-wait

	h.Close()

	err := <-h.errs
	if err != ErrSome {
		t.Fatalf("Expected ErrSome, got %+v", err)
	}

}
