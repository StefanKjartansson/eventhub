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
}

func NewHub(application string, d DataBackend) *hub {
	return &hub{
		application: application,
		db:          d,
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

func (h *hub) AddDataService(d DataService) {
	h.m.Lock()
	defer h.m.Unlock()
	d.SetBackend(&h.db)
	h.dataservices = append(h.dataservices, d)
}

func (h *hub) Run() {

	merged := Merge(h.feeds...)

	for {
		var e *Event
		select {
		case e = <-merged.Updates():
			err := h.db.Save(e)
			if err != nil {
				err = merged.Close()
			}
			for _, b := range h.broadcasters {
				b.Broadcast(e)
			}
		}
	}

}
