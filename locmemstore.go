package eventhub

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Events []*Event

func (e Events) Len() int      { return len(e) }
func (e Events) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

type ByDate struct{ Events }

func (s ByDate) Less(i, j int) bool {
	return s.Events[i].Updated.Nanosecond() > s.Events[j].Updated.Nanosecond()
}

type LocalMemoryStore struct {
	evs []Event
	m   sync.Mutex
}

func (d *LocalMemoryStore) GetById(id int) (*Event, error) {

	d.m.Lock()
	defer d.m.Unlock()

	for _, e := range d.evs {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, errors.New("No event found")
}

func (d *LocalMemoryStore) Save(e *Event) error {

	d.m.Lock()
	defer d.m.Unlock()

	if e.ID != 0 {
		switchIdx := -1
		for idx, x := range d.evs {
			if x.ID == e.ID {
				switchIdx = idx
			}
		}
		if switchIdx == -1 {
			return errors.New("ID provided but no event found with provided id")
		}
		e.Updated = time.Now()
		d.evs[switchIdx] = *e
	} else {
		e.ID = len(d.evs) + 1
		t := time.Now()
		e.Created = t
		e.Updated = t
		d.evs = append(d.evs, *e)
	}

	return nil
}

func (d *LocalMemoryStore) Query(q Query) ([]*Event, error) {

	d.m.Lock()
	defer d.m.Unlock()

	var matched []*Event

	for idx, event := range d.evs {
		if q.Match(event) {
			matched = append(matched, &d.evs[idx])
		}
	}

	sort.Sort(ByDate{matched})

	return matched, nil
}

func (d *LocalMemoryStore) AggregateType(q Query, s string) (map[string]int, error) {

	if !stringInSlice(s, []string{"entities", "other_references", "actors", "tags"}) {
		return nil, errors.New("Invalid type")
	}

	m := make(map[string]int)

	evs, err := d.Query(q)

	if err != nil {
		return m, err
	}

	sumArray := func(arr []string) {
		for _, i := range arr {
			_, ok := m[i]
			if !ok {
				m[i] = 1
			} else {
				m[i]++
			}
		}
	}

	for _, e := range evs {
		switch s {
		case "entities":
			sumArray(e.Entities)
		case "other_references":
			sumArray(e.OtherReferences)
		case "actors":
			sumArray(e.Actors)
		case "tags":
			sumArray(e.Tags)
		}
	}

	return m, nil
}

func (d *LocalMemoryStore) Clear() {
	d.m.Lock()
	defer d.m.Unlock()
	d.evs = nil
}

func NewLocalMemoryStore() *LocalMemoryStore {
	d := LocalMemoryStore{}
	return &d
}
