package db

import (
	"database/sql"
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"log"
)

type PostgresDataSource struct {
	pg *sql.DB
}

//Converts a row to an event
func scanRow(row *sql.Rows, e *eventhub.Event) error {

	var entities StringSlice
	var references StringSlice
	var actors StringSlice
	var tags StringSlice
	temp := []byte{}

	err := row.Scan(
		&e.ID,
		&e.Key,
		&e.Created,
		&e.Updated,
		&temp,
		&e.Description,
		&e.Importance,
		&e.Origin,
		&entities,
		&references,
		&actors,
		&tags)

	if err != nil {
		return err
	}

	var data interface{}
	err = json.Unmarshal(temp, &data)

	if err != nil {
		return err
	}

	e.Payload = data
	e.Entities = entities
	e.OtherReferences = references
	e.Actors = actors
	e.Tags = tags

	return nil
}

//Gets an event by id
func (p *PostgresDataSource) GetById(id int) (*eventhub.Event, error) {

	var e eventhub.Event
	log.Println(id)
	err := wrapTransaction(p.pg, func(tx *sql.Tx) error {
		rows, err := tx.Query(`
        SELECT
            *
        FROM
            "event"
        WHERE "id" = $1
        `, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return sql.ErrNoRows
		}
		return scanRow(rows, &e)
	})

	if err != nil {
		return nil, err
	}
	return &e, nil

}

//Creates a new PostgresDataSource
func NewPostgresDataSource(connection string) (*PostgresDataSource, error) {

	p := PostgresDataSource{}

	pg, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	p.pg = pg

	//Runs migrations
	bootstrapDatabase(p.pg)
	return &p, nil
}
