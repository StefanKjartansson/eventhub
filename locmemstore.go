package eventhub

import (
	"errors"
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

type LocalMemoryStore struct {
	evs []Event
	m   sync.Mutex
	ch  chan *Event
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

	d.ch <- e

	return nil
}

func (d *LocalMemoryStore) Updates() <-chan *Event {
	return d.ch
}

func (d *LocalMemoryStore) Close() error {
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

func (d *LocalMemoryStore) FilterBy(m map[string]interface{}) ([]*Event, error) {

	d.m.Lock()
	defer d.m.Unlock()

	var matched []*Event

	for idx, event := range d.evs {

		r := reflect.ValueOf(event)
		if r.Kind() == reflect.Ptr {
			r = r.Elem()
		}
		match := false
		for key, value := range m {
			f := r.FieldByName(key)

			if reflect.DeepEqual(f.Interface(), value) {
				match = true
			}

			//field is array
			fieldValueAsArray, fieldOk := f.Interface().([]string)
			vAsArray, ok := value.([]string)

			//both are string arrays
			if ok && fieldOk {
				eventData := fieldValueAsArray
				allMatch := true
				for _, s := range vAsArray {
					if !stringInSlice(s, eventData) {
						allMatch = false
					}
				}
				match = allMatch
			}

			//query field is array and field is a string
			if !fieldOk && ok {
				anyMatch := false
				asString := f.Interface().(string)
				for _, s := range vAsArray {
					if anyMatch {
						continue
					}
					if !anyMatch && asString == s {
						anyMatch = true
					}
				}
				match = anyMatch
			}

		}
		if match {
			matched = append(matched, &d.evs[idx])
		}
	}

	sort.Sort(ByDate{matched})

	return matched, nil
}

func (d *LocalMemoryStore) Clear() {
	d.m.Lock()
	defer d.m.Unlock()
	d.evs = nil
}

func NewLocalMemoryStore() *LocalMemoryStore {
	d := LocalMemoryStore{}
	d.ch = make(chan *Event)
	return &d
}
