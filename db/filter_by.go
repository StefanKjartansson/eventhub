package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"log"
	"reflect"
	"strings"
)

func (p *PostgresDataSource) FilterBy(m map[string]interface{}) ([]*eventhub.Event, error) {

	events := []*eventhub.Event{}
	var buffer bytes.Buffer

	buffer.WriteString("select * from event where ")

	r := reflect.ValueOf(eventhub.Event{})
	args := []interface{}{}
	cnt := 1
	paramCount := 1
	l := len(m)

	for k, v := range m {

		f := r.FieldByName(k)
		k = strings.ToLower(k)

		switch v.(type) {

		case []string:

			sArray := v.([]string)

			// attempt to type cast the field value as []string
			_, fieldOk := f.Interface().([]string)

			if fieldOk {
				//db field is array.
				if len(sArray) == 1 {
					buffer.WriteString(fmt.Sprintf("$%d = ANY(%s)", paramCount, k))
					args = append(args, sArray[0])
					paramCount++
				} else {
					buffer.WriteString(fmt.Sprintf("%s @> ARRAY[", k))
					arrLen := len(sArray)
					for arrIdx, i := range sArray {
						buffer.WriteString(fmt.Sprintf("$%d", paramCount))

						if arrIdx+1 < arrLen {
							buffer.WriteString(", ")
						}

						args = append(args, i)
						paramCount++
					}
					buffer.WriteString("]::text[]")
				}
			} else {
				arrLen := len(sArray)
				for arrIdx, i := range sArray {
					buffer.WriteString(fmt.Sprintf("%s = $%d", k, paramCount))
					if arrIdx+1 < arrLen {
						buffer.WriteString(" or ")
					}
					args = append(args, i)
					paramCount++
				}
			}

		default:
			buffer.WriteString(fmt.Sprintf(`%s = $%d`, k, cnt))
			args = append(args, v)
			paramCount++

		}

		if cnt < l {
			buffer.WriteString(" and ")
		} else {
			buffer.WriteString(" order by updated desc;")
		}
		cnt++

	}
	log.Println(buffer.String())
	log.Println(args)
	err := wrapTransaction(p.pg, func(tx *sql.Tx) error {
		rows, err := tx.Query(buffer.String(), args...)
		defer rows.Close()
		for rows.Next() {
			var e eventhub.Event
			err = scanRow(rows, &e)
			if err != nil {
				log.Println(err)
				return err
			}
			events = append(events, &e)
		}
		return err
	})
	log.Println(len(events), err)

	return events, err
}
