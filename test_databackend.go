package eventhub

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func bootstrapData(d DataBackend) {

	r := rand.New(rand.NewSource(5))

	keys := make(map[int]string)
	keys[0] = "bar"
	keys[1] = "baz"
	keys[2] = "boo"

	actors := make(map[int][]string)
	actors[0] = []string{"employee1"}
	actors[1] = []string{"employee1", ""}
	actors[2] = []string{"employee1"}
	actors[3] = []string{"employee1"}
	actors[4] = []string{"employee1"}

	for i := 0; i < 20; i++ {

		actors := []string{}
		for j := 0; j < i; j++ {
			actors = append(actors, fmt.Sprintf("employee%d", j))
		}

		_ = d.Save(NewEvent(
			"foo."+keys[r.Intn(len(keys))],
			nil,
			nil,
			"ba ba",
			r.Intn(5),
			"mysystem",
			[]string{"ns/foo", "ns/moo"},
			nil,
			actors,
			nil))

	}

}

func FilterByTest(t *testing.T, d DataBackend) {

	bootstrapData(d)

	q := Query{}
	q.Origin = "mysystem"

	evs, err := d.Query(q)

	if err != nil {
		t.Fatal(err)
	}

	if len(evs) != 20 {
		t.Fatal("Filter results count should be 20")
	}

	if evs[0].Origin != "mysystem" {
		t.Fatal("Origin not expected")
	}

	q = Query{}

	for i := 0; i < 20; i++ {

		expected := 20 - i
		for j := 0; j < i; j++ {
			q.Actors = append(q.Actors, fmt.Sprintf("employee%d", j))
		}

		if len(q.Actors) == 0 {
			continue
		}

		evs, err = d.Query(q)

		if err != nil {
			t.Fatal(err)
		}

		if len(evs) != expected {
			t.Fatalf("Expected %d, got %d. query: %+v",
				expected, len(evs), q)
		}

	}

	q = Query{}
	q.Key = "foo.bar OR foo.baz OR foo.boo"

	evs, err = d.Query(q)

	if err != nil {
		t.Fatal(err)
	}

	if len(evs) != 20 {
		t.Fatalf("evs returned %d, expected 20, query:%+v", len(evs), q)
	}

	//Test that ordering is correct
	var ti time.Time
	for idx, e := range evs {
		if idx == 0 {
			ti = e.Updated
			continue
		}
		// newest one should be first
		if e.Updated.UnixNano() > ti.UnixNano() {
			t.Fatalf("Expected events to have been ordering in descending order: %s, got %s", ti, e.Updated)
		}
		ti = e.Updated
	}
}

func InsertUpdateTest(t *testing.T, d DataBackend) {

	data := struct {
		Foo string `json:"foo"`
	}{
		"bar",
	}

	e := NewEvent(
		"foo.bar",
		nil,
		data,
		"My event",
		3,
		"mysystem",
		[]string{"ns/foo", "ns/moo"},
		nil,
		[]string{"someone"},
		nil)

	err := d.Save(e)
	if err != nil {
		t.Fatal(err)
	}

	if e.ID != 1 {
		t.Fatal("Expected '%d', got %v", 1, e.ID)
	}

	newE, err := d.GetById(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", newE)

	newE.Description = "New Description"
	err = d.Save(newE)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := d.GetById(1)
	if updated.Description != newE.Description {
		t.Fatal(err)
	}

	err = d.Save(NewEvent(
		"foo.baz",
		nil,
		data,
		"My event 2",
		3,
		"mysystem",
		[]string{"ns/foo", "ns/moo"},
		nil,
		[]string{"someone"},
		nil))

	if err != nil {
		t.Fatal(err)
	}
}
