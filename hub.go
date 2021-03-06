package straumur

import (
	"github.com/howbazaar/loggo"
)

type hub struct {
	feeds        []EventFeed
	db           DataBackend
	broadcasters []Broadcaster
	dataservices []DataService
	Errs         chan error
	quit         chan struct{}
	processors   map[string]*processorList
}

var (
	logger = loggo.GetLogger("straumur.core")
)

func NewHub(d DataBackend) *hub {
	return &hub{
		db:         d,
		Errs:       make(chan error),
		quit:       make(chan struct{}),
		processors: make(map[string]*processorList),
	}
}

func (h *hub) RegisterProcessor(step, pattern string, f Processor) error {
	pl, ok := h.processors[step]
	if !ok {
		pl = NewProcessorList()
		h.processors[step] = pl
	}
	err := pl.Register(pattern, f)
	if err != nil {
		logger.Errorf("Registering processor failed: %+v", err)
	}
	return err
}

func (h *hub) AddFeeds(efs ...EventFeed) {
	for _, e := range efs {
		h.feeds = append(h.feeds, e)
	}
}

func (h *hub) AddBroadcasters(bcs ...Broadcaster) {
	for _, b := range bcs {
		h.broadcasters = append(h.broadcasters, b)
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
				h.Errs <- pl.Process(e)
			}
			err := h.db.Save(e)
			if err != nil {
				h.Errs <- err
			}
			for _, b := range h.broadcasters {
				go b.Broadcast(e)
			}

		case <-h.quit:
			h.Errs <- merged.Close()
			return
		}
	}

	return
}

func (h *hub) Close() {
	logger.Infof("Closing hub")
	close(h.quit)
}
