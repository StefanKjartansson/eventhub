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

	d.Save(&e)

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
}
