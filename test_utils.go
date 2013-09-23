package eventhub

import (
	"fmt"
	"math/rand"
	"testing"
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

		_ = d.Save(&Event{
			Key:         "foo." + keys[r.Intn(len(keys))],
			Description: "ba ba",
			Importance:  r.Intn(5),
			Origin:      "mysystem",
			Entities:    []string{"ns/foo", "ns/moo"},
			Actors:      actors,
		})
	}

}

func FilterByTest(t *testing.T, d DataBackend) {

	bootstrapData(d)

	m := make(map[string]interface{})
	m["Origin"] = "mysystem"

	evs, err := d.FilterBy(m)

	if err != nil {
		t.Fatal(err)
	}

	if len(evs) != 20 {
		t.Fatal("Filter results count should be 20")
	}

	if evs[0].Origin != "mysystem" {
		t.Fatal("Origin not expected")
	}

	delete(m, "Origin")

	for i := 0; i < 20; i++ {

		expected := 20 - i
		actors := []string{}
		for j := 0; j < i; j++ {
			actors = append(actors, fmt.Sprintf("employee%d", j))
		}
		m["Actors"] = actors

		if len(actors) == 0 {
			continue
		}

		evs, err = d.FilterBy(m)

		if err != nil {
			t.Fatal(err)
		}

		if len(evs) != expected {
			t.Fatalf("Expected %d, got %d. actors: %v", expected, len(evs), actors)
		}

	}

}

func InsertUpdateTest(t *testing.T, d DataBackend) {

	data := struct {
		Foo string `json:"foo"`
	}{
		"bar",
	}

	e := Event{
		Key:         "foo.bar",
		Payload:     data,
		Description: "My event",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
		Actors:      []string{"someone"},
	}

	err := d.Save(&e)
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

	err = d.Save(&Event{
		Key:         "foo.baz",
		Payload:     data,
		Description: "My event 2",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
		Actors:      []string{"someone"},
	})

	if err != nil {
		t.Fatal(err)
	}
}
