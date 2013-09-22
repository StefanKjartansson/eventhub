package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"strings"
)

func (p *PostgresDataSource) FilterBy(m map[string]interface{}) ([]*eventhub.Event, error) {

	events := []*eventhub.Event{}
	var buffer bytes.Buffer

	buffer.WriteString("select * from event where ")

	args := []interface{}{}
	cnt := 1
	l := len(m)
	for k, v := range m {
		buffer.WriteString(fmt.Sprintf(`%s = $%d`, strings.ToLower(k), cnt))
		args = append(args, v)
		if cnt < l {
			buffer.WriteString(" and ")
		} else {
			buffer.WriteString(";")
		}
		cnt++
	}

	stmt, err := p.pg.Prepare(buffer.String())
	if err != nil {
		return nil, err
	}

	err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
		rows, err := tx.Stmt(stmt).Query(args...)
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

	return events, nil
}
