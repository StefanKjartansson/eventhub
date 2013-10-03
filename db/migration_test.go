package db

import (
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
