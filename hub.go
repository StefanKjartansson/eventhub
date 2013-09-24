package eventhub

type merge struct {
	feeds   []EventFeed
	updates chan Event
	quit    chan struct{}
	errs    chan error
}

func Merge(feeds ...EventFeed) EventFeed {
	m := &merge{
		feeds:   feeds,
		updates: make(chan Event),
		quit:    make(chan struct{}),
		errs:    make(chan error),
	}
	for _, feed := range feeds {
		go func(f EventFeed) {
			for {
				var it Event
				select {
				case it = <-f.Updates():
				case <-m.quit: // HL
					m.errs <- f.Close() // HL
					return              // HL
				}
				select {
				case m.updates <- it:
				case <-m.quit: // HL
					m.errs <- f.Close() // HL
					return              // HL
				}
			}
		}(feed)
	}
	return m
}

func (m *merge) Updates() <-chan Event {
	return m.updates
}

func (m *merge) Close() (err error) {
	close(m.quit) // HL
	for _ = range m.feeds {
		if e := <-m.errs; e != nil { // HL
			err = e
		}
	}
	close(m.updates) // HL
	return
}
