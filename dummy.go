package eventhub

import (
	"errors"
	"sync"
)

type DummyDataSource struct {
	evs []*Event
	m   sync.Mutex
}

func (d *DummyDataSource) GetById(id int) (*Event, error) {

	d.m.Lock()
	defer d.m.Unlock()

	for _, e := range d.evs {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, errors.New("No event found")
}

func (d *DummyDataSource) Save(e *Event) error {

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
		d.evs[switchIdx] = e
	} else {
		e.ID = len(d.evs) + 1
		d.evs = append(d.evs, e)
	}

	return nil
}

func (d *DummyDataSource) FilterBy(m map[string]interface{}) ([]*Event, error) {
	d.m.Lock()
	defer d.m.Unlock()
	var matched []*Event
	matchedIndexes := []int{}

	for idx, event := range d.evs {
		match := false
		for key, value := range m {
			//TODO, get field at reflect and DeepEqual 
		}
		if match {
			matchedIndexes = append(matchedIndexes, idx)
		}
	}

	return matched, nil
}

func (d *DummyDataSource) Clear() {
	d.m.Lock()
	defer d.m.Unlock()
	d.evs = nil
}

func NewDummyBackend() *DummyDataSource {
	return &DummyDataSource{}
}
