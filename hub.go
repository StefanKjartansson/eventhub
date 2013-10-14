package straumur

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
	quit         chan struct{}
	processors   map[string]*processorList
}

func NewHub(application string, d DataBackend) *hub {
	return &hub{
		application: application,
		db:          d,
		errs:        make(chan error),
		quit:        make(chan struct{}),
		processors:  make(map[string]*processorList),
	}
}

func (h *hub) RegisterProcessor(step, pattern string, f Processor) error {
	pl, ok := h.processors[step]
	if !ok {
		pl = NewProcessorList()
		h.processors[step] = pl
	}
	return pl.Register(pattern, f)
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
			pl, ok := h.processors["pre"]
			if ok {
				h.errs <- pl.Process(e)
			}
			err := h.db.Save(e)
			if err != nil {
				h.errs <- err
			}
			for _, b := range h.broadcasters {
				go b.Broadcast(e)
			}

		case <-h.quit:
			h.errs <- merged.Close()
			return
		}
	}

	return
}

func (h *hub) Close() {
	close(h.quit)
}
