package db

import (
	"database/sql"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"testing"
)

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

	eventhub.InsertUpdateTest(t, p)

	//Clear the table
	_, err = db.Exec(`truncate table event;`)

	eventhub.FilterByTest(t, p)
}
