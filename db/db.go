package db

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
	"strings"
    "github.com/StefanKjartansson/eventhub"
)

const (
	insertStatement = `
    INSERT INTO
        "event"
        ("key", "created", "payload", "description", "importance", "origin",
         "entities", "other_references", "actors", "tags")
    VALUES
        ($1, $2, $3, $4, $5, $6, ARRAY[$7], ARRAY[$8], ARRAY[$9], ARRAY[$10])
    RETURNING "id";
    `

    selectById = `
    SELECT "id", "key", "created", "payload", "description", "importance",
        "origin", "entities", "other_references", "actors", "tags"
    FROM
        "event"
    WHERE "id" = $1
    `
)

type PostgresDataSource struct {
	pg         *sql.DB
	insert     *sql.Stmt
	selectbyid *sql.Stmt
}

func (p *PostgresDataSource) GetById(id int) (*eventhub.Event, error) {

	var e eventhub.Event
	var err error
	var tx *sql.Tx

	if tx, err = p.pg.Begin(); err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()
		}
	}()

	var entities StringSlice
	var references StringSlice
	var actors StringSlice
	var tags StringSlice

	err = tx.Stmt(p.selectbyid).QueryRow(id).Scan(
		&e.ID,
		&e.Key,
		&e.Created,
		&e.Payload,
		&e.Description,
		&e.Importance,
		&e.Origin,
		&entities,
		&references,
		&actors,
		&tags)

	if err != nil {
		return nil, err
	}

	e.Entities = entities
	e.OtherReferences = references
	e.Actors = actors
	e.Tags = tags

	return &e, nil

}

func (p *PostgresDataSource) Save(e *eventhub.Event) (err error) {

	var tx *sql.Tx

	if tx, err = p.pg.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()
		}
	}()

	b, err := json.Marshal(e.Payload)

	err = tx.Stmt(p.insert).QueryRow(
		e.Key,
		e.Created,
		b,
		e.Description,
		e.Importance,
		e.Origin,
		strings.Join(e.Entities, ", "),
		strings.Join(e.OtherReferences, ", "),
		strings.Join(e.Actors, ", "),
		strings.Join(e.Tags, ", ")).Scan(&e.ID)
	if err != nil {
		return err
	}

	return nil
}

func NewPostgresDataSource(connection string) (*PostgresDataSource, error) {

	p := PostgresDataSource{}

	pg, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	p.pg = pg

	bootstrapDatabase(p.pg)

	s, err := pg.Prepare(insertStatement)
	if err != nil {
		return &p, err
	}
	p.insert = s

	s, err = pg.Prepare(selectById)
	if err != nil {
		return &p, err
	}
	p.selectbyid = s

	return &p, nil
}
