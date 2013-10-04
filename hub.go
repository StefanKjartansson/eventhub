package eventhub

import (
	"log"
	"sync"
)

type hub struct {
	application  string
	m            sync.Mutex
	feeds        []EventFeed
	db           DataBackend
	broadcasters []Broadcaster
	dataservices []DataService
	errs         chan error
	quit         chan struct{}
}

func NewHub(application string, d DataBackend) *hub {
	return &hub{
		application: application,
		db:          d,
		errs:        make(chan error),
		quit:        make(chan struct{}),
	}
}

func (h *hub) AddFeeds(efs ...EventFeed) {
	h.m.Lock()
	defer h.m.Unlock()
	for _, e := range efs {
		h.feeds = append(h.feeds, e)
	}
}

func (h *hub) AddBroadcasters(bcs ...Broadcaster) {
	h.m.Lock()
	defer h.m.Unlock()
	for _, b := range bcs {
		h.broadcasters = append(h.broadcasters, b)
	}
}

func (h *hub) AddDataServices(ds ...DataService) {
	for _, d := range ds {
		go d.Run(h.db, h.errs)
	}
}

func (h *hub) Run() {

	merged := Merge(h.feeds...)

	for {
		var e *Event
		select {
		case e = <-merged.Updates():
			err := h.db.Save(e)
			if err != nil {
				h.errs <- err
			}
			for _, b := range h.broadcasters {
				go b.Broadcast(e)
			}
		case err := <-h.errs:
			log.Fatal(err)
			return
		case <-h.quit:
			h.errs <- merged.Close()
			return
		}
	}

}

func (h *hub) Close() error {
	close(h.quit)
	return <-h.errs
}
