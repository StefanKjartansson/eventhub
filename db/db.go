package db

import (
	"database/sql"
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"strings"
)

type PostgresDataSource struct {
	pg         *sql.DB
	insert     *sql.Stmt
	update     *sql.Stmt
	selectbyid *sql.Stmt
}

func scanRow(row *sql.Rows, e *eventhub.Event) error {
	var err error
	var entities StringSlice
	var references StringSlice
	var actors StringSlice
	var tags StringSlice
	temp := []byte{}

	err = row.Scan(
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

func (p *PostgresDataSource) GetById(id int) (*eventhub.Event, error) {

	var e eventhub.Event
	var err error

	err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
		rows, err := tx.Stmt(p.selectbyid).Query(id)
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

func (p *PostgresDataSource) Save(e *eventhub.Event) (err error) {

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
		err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
			return tx.Stmt(p.insert).QueryRow(args...).Scan(&e.ID, &e.Created, &e.Updated)
		})
	default:
		args := append(args, e.ID)
		err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
			return tx.Stmt(p.update).QueryRow(args...).Scan(&e.Updated)
		})
	}

	return err
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
