package eventstream

import (
	"database/sql"
	_ "github.com/lib/pq"
	"testing"
	"time"
)

func TestMigrationsGlob(t *testing.T) {

	const expected = "migrations/migrations-2013-09-20T09-01-45Z.sql"
	expectedDate := time.Date(2013, time.September, 20, 9, 1, 45, 0, time.UTC)

	m, err := globMigrations()
	if err != nil {
		t.Error("glob has error:", err)
		return
	}

	t.Logf("%v", m)

	if m[0].filename != expected {
		t.Errorf("Expected '%s', got %s", expected, m[0].filename)
		return
	}

	if m[0].date != expectedDate {
		t.Errorf("Expected '%s', got %s", expectedDate, m[0].date)
		return
	}

	originalLength := len(m)
	m = m.FilterDates([]time.Time{expectedDate})
	if originalLength == len(m) {
		t.Errorf("Expected length: %d, got %d", originalLength-1, originalLength)
		return
	}
}

func TestDB(t *testing.T) {

	const connection = "dbname=teststream host=localhost sslmode=disable"

	db, err := sql.Open("postgres", connection)
	if err != nil {
		t.Error("Error:", err)
		return
	}

	_, err = db.Exec(`drop table if exists migration_info, event;`)
	if err != nil {
		t.Error("Error:", err)
		return
	}

	// With migrations applied
	_, err = NewPostgresDataSource(connection)
	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	// With no migrations applied
	p, err := NewPostgresDataSource(connection)
	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	data := struct {
		Foo string
	}{
		"bar",
	}

	e := Event{
		Key:         "foo.bar",
		Created:     time.Now(),
		Payload:     data,
		Description: "My event",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
		Actors:      []string{"someone"},
	}

	p.Insert(&e)

	if e.ID != 1 {
		t.Errorf("Expected '%d', got %v", 1, e.ID)
	}

	newE, err := p.GetById(1)

	if err != nil {
		t.Error("PostgresDataSource has error:", err)
		return
	}

	t.Logf("%v", newE)
}
