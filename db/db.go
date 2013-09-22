package db

import (
	"database/sql"
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

const (
	insertSQL = `
    INSERT INTO "event"
        (
            "key",
            "created",
            "updated",
            "payload",
            "description",
            "importance",
            "origin",
            "entities",
            "other_references",
            "actors",
            "tags"
        )
    VALUES
        (
            $1,
            now(),
            now(),
            $2,
            $3,
            $4,
            $5,
            ARRAY[$6],
            ARRAY[$7],
            ARRAY[$8],
            ARRAY[$9]
        )
    RETURNING
        "id",
        "created",
        "updated";
    `

	updateSQL = `
    UPDATE "event"
    SET
        "key" = $1,
        "payload" = $2,
        "description" = $3,
        "importance" = $4,
        "origin" = $5,
        "entities" = ARRAY[$6],
        "other_references" = ARRAY[$7],
        "actors" = ARRAY[$8],
        "tags" = ARRAY[$9],
        "updated" = now()
    WHERE
        "id" = $10
    RETURNING
        "updated";
    `

	selectByIdSQL = `
    SELECT
        "id",
        "key",
        "created",
        "updated",
        "payload",
        "description",
        "importance",
        "origin",
        "entities",
        "other_references",
        "actors",
        "tags"
    FROM
        "event"
    WHERE "id" = $1
    `
)

type PostgresDataSource struct {
	pg         *sql.DB
	insert     *sql.Stmt
	update     *sql.Stmt
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

	temp := []byte{}

	err = tx.Stmt(p.selectbyid).QueryRow(id).Scan(
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
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(temp, &data)

	if err != nil {
		return nil, err
	}

	e.Payload = data
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
	if err != nil {
		return err
	}

	args := []interface{}{
		e.Key,
		b,
		e.Description,
		e.Importance,
		e.Origin,
		strings.Join(e.Entities, ", "),
		strings.Join(e.OtherReferences, ", "),
		strings.Join(e.Actors, ", "),
		strings.Join(e.Tags, ", "),
	}

	switch e.ID {
	case 0:
		err = tx.Stmt(p.insert).QueryRow(args...).Scan(&e.ID, &e.Created, &e.Updated)
	default:
		args := append(args, e.ID)
		err = tx.Stmt(p.update).QueryRow(args...).Scan(&e.Updated)
	}

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

	s, err := pg.Prepare(insertSQL)
	if err != nil {
		return &p, err
	}
	p.insert = s

	s, err = pg.Prepare(selectByIdSQL)
	if err != nil {
		return &p, err
	}
	p.selectbyid = s

	s, err = pg.Prepare(updateSQL)
	if err != nil {
		return &p, err
	}
	p.update = s

	return &p, nil
}
