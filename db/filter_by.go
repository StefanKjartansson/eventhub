package db

import (
	"database/sql"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
)

func (p *PostgresDataSource) Query(q eventhub.Query) ([]*eventhub.Event, error) {

	events := []*eventhub.Event{}

	query, args := buildQuery(q)

	err := wrapTransaction(p.pg, func(tx *sql.Tx) error {
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
