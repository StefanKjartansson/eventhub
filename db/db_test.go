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

	//truncate table
	clear := func() {
		_, err := db.Exec(`TRUNCATE TABLE event RESTART IDENTITY;`)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = db.Exec(`drop table if exists migration_info, event;`)
	if err != nil {
		t.Fatal("Error:", err)
	}

	// With migrations applied
	_, err = NewPostgresDataSource(connection)
	if err != nil {
		t.Fatal("PostgresDataSource has error:", err)
	}

	// With no migrations applied
	p, err := NewPostgresDataSource(connection)
	if err != nil {
		t.Fatal("PostgresDataSource has error:", err)
	}

	eventhub.RunDataBackendSuite(t, p, clear)
}
