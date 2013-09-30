package eventhub

import (
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
}

func NewHub(application string, d DataBackend) *hub {
	return &hub{
		application: application,
		db:          d,
		errs:        make(chan error),
	}
}

func (h *hub) AddFeed(e EventFeed) {
	h.m.Lock()
	defer h.m.Unlock()
	h.feeds = append(h.feeds, e)
}

func (h *hub) AddBroadcaster(b Broadcaster) {
	h.m.Lock()
	defer h.m.Unlock()
	h.broadcasters = append(h.broadcasters, b)
}

func (h *hub) AddDataService(d DataService) {
	h.m.Lock()
	defer h.m.Unlock()
	h.errs <- d.SetBackend(&h.db)
	h.dataservices = append(h.dataservices, d)
}

func (h *hub) Run() {

	merged := Merge(h.feeds...)

	for _, ds := range h.dataservices {
		go func(d DataService) {
			h.errs <- ds.Run()
		}(ds)
	}

	for {
		var e *Event
		select {
		case e = <-merged.Updates():
			h.errs <- h.db.Save(e)
			for _, bc := range h.broadcasters {
				go func(b Broadcaster) {
					h.errs <- b.Broadcast(e)
				}(bc)
			}
		}
	}
}
