package eventhub

import (
	"errors"
	"log"
	"reflect"
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

type DummyDataSource struct {
	evs Events
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
		e.Updated = time.Now()
		d.evs[switchIdx] = e
	} else {
		e.ID = len(d.evs) + 1
		t := time.Now()
		e.Created = t
		e.Updated = t
		d.evs = append(d.evs, e)
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (d *DummyDataSource) FilterBy(m map[string]interface{}) ([]*Event, error) {
	d.m.Lock()
	defer d.m.Unlock()
	var matched Events

	for _, event := range d.evs {
		r := reflect.ValueOf(event)
		if r.Kind() == reflect.Ptr {
			r = r.Elem()
		}
		match := false
		for key, value := range m {
			f := r.FieldByName(key)
			//TODO: Emulate ANY
			log.Printf("field value: %v, query value: %v", f.Interface(), value)
			if reflect.DeepEqual(f.Interface(), value) {
				match = true
			}
			if vAsArray, ok := value.([]string); ok {
				eventData := f.Interface().([]string)
				allMatch := true
				for _, s := range vAsArray {
					if !stringInSlice(s, eventData) {
						allMatch = false
					}
				}
				match = allMatch
			}
		}
		if match {
			matched = append(matched, event)
		}
	}

	sort.Sort(ByDate{matched})

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
