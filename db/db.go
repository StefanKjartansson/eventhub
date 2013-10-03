package db

import (
	"database/sql"
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"log"
	"time"
)

// Callback for a managed transaction
//
// Example:
//
//	err := p.wrapTransaction(func(tx *sql.Tx) error {
//	    rows, err := tx.Query(query, args...)
//      if err != nil {
//         return err
//      }
// }
//
type TransactionFunc func(*sql.Tx) error

type PostgresDataSource struct {
	pg *sql.DB
}

// Converts a row to an event
func scanRow(row *sql.Rows, e *eventhub.Event) error {

	var entities StringSlice
	var references StringSlice
	var actors StringSlice
	var tags StringSlice
	temp := []byte{}
	tempkey := []byte{}

	err := row.Scan(
		&e.ID,
		&e.Key,
		&tempkey,
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

	var keydata interface{}
	err = json.Unmarshal(tempkey, &keydata)

	if err != nil {
		return err
	}

	e.Payload = data
	e.KeyParams = keydata
	e.Entities = entities
	e.OtherReferences = references
	e.Actors = actors
	e.Tags = tags

	return nil
}

//Gets an event by id
func (p *PostgresDataSource) GetById(id int) (*eventhub.Event, error) {

	var e eventhub.Event

	err := p.wrapTransaction(func(tx *sql.Tx) error {
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

func (d *PostgresDataSource) applyMigrations() {

	// Get all table names
	// TODO: maybe change the schema name?

	rows, err := d.pg.Query(`
        select tablename
            from pg_tables
        where
            pg_tables.schemaname = 'public';
    `)

	if err != nil {
		log.Fatal(err)
	}

	canMigrate := false
	var s string
	for rows.Next() {
		rows.Scan(&s)
		if s == "migration_info" {
			canMigrate = true
		}
	}

	//No table names returned
	if s == "" {
		canMigrate = true
	}

	//Get the list of migrations
	m, err := globMigrations()

	if err != nil {
		log.Fatal(err)
	}

	//If there were tables, the migration_info
	//table should be among them
	if s != "" {
		rows, err := d.pg.Query(`
            select created from
                migration_info
            order by created
        `)

		removalDates := []time.Time{}
		for rows.Next() {
			var t time.Time
			err = rows.Scan(&t)
			if err != nil {
				log.Fatal(err)
			}
			//Weird, table created with TZ, but Scan doesn't
			//add the UTC info
			removalDates = append(removalDates, t.UTC())
		}

		//Filter out migrations which have already been applied
		m = m.FilterDates(removalDates)
	}

	//Run migrations
	if canMigrate && len(m) > 0 {

		for _, migration := range m {

			_, err := d.pg.Exec(migration.content)
			if err != nil {
				log.Fatal(err)
			}
			_, err = d.pg.Exec(`
                insert into migration_info
                    (created, content)
                values($1, $2)`, migration.date, migration.content)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}

func (d *PostgresDataSource) wrapTransaction(t TransactionFunc) (err error) {

	var tx *sql.Tx

	if tx, err = d.pg.Begin(); err != nil {
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

	return t(tx)

}

func (p *PostgresDataSource) Query(q eventhub.Query) ([]*eventhub.Event, error) {

	events := []*eventhub.Event{}

	query, args := buildSelectQuery(q)

	err := p.wrapTransaction(func(tx *sql.Tx) error {
		rows, err := tx.Query(query, args...)
		defer rows.Close()
		for rows.Next() {
			var e eventhub.Event
			err = scanRow(rows, &e)
			if err != nil {
				return err
			}
			events = append(events, &e)
		}
		return err
	})

	return events, err
}

// Saves or updates an event
func (p *PostgresDataSource) Save(e *eventhub.Event) (err error) {

	switch e.ID {
	case 0:
		err = p.wrapTransaction(func(tx *sql.Tx) error {
			query, args, err := buildInsertQuery(e)
			if err != nil {
				return err
			}
			return tx.QueryRow(query, args...).Scan(&e.ID, &e.Created, &e.Updated)
		})
	default:
		err = p.wrapTransaction(func(tx *sql.Tx) error {
			query, args, err := buildUpdateQuery(e)
			if err != nil {
				return err
			}
			return tx.QueryRow(query, args...).Scan(&e.Updated)
		})
	}

	return err
}

//Creates a new PostgresDataSource
func NewPostgresDataSource(connection string) (*PostgresDataSource, error) {

	p := PostgresDataSource{}

	pg, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	p.pg = pg

	//Run migrations
	p.applyMigrations()

	return &p, nil
}
