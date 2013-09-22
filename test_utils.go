package eventhub

import (
	"testing"
)

func RunDataBackendTest(t *testing.T, d DataBackend) {

	data := struct {
		Foo string
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
		t.Error("PostgresDataSource has error:", err)
		return
	}

	if e.ID != 1 {
		t.Errorf("Expected '%d', got %v", 1, e.ID)
	}

	newE, err := d.GetById(1)
	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	t.Logf("%v", newE)

	newE.Description = "New Description"
	err = d.Save(newE)
	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	updated, err := d.GetById(1)
	if updated.Description != newE.Description {
		t.Error("PostgresDataSource has error:", err)
		return
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
		t.Error("PostgresDataSource has error:", err)
		return
	}

	m := make(map[string]interface{})
	//m["Key"] = "foo.*"
	m["Origin"] = "mysystem"

	evs, err := d.FilterBy(m)

	t.Log(evs)

	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	if len(evs) != 2 {
		t.Error("Filter results count should've been 2")
		return
	}

	if evs[0].Origin != "mysystem" {
		t.Error("Origin not expected")
		return
	}
}
