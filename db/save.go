package db

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
)

func writeArrayParams(buffer *bytes.Buffer, arrays [][]string, startParam int, keys []string) int {

	//todo, accept more than just string arrays
	paramCount := startParam
	arrLength := len(arrays)
	idx := 0

	writeKey := false
	if keys != nil {
		writeKey = true
		if len(keys) != len(arrays) {
			panic("Illegal")
		}
	}

	for keyIdx, arr := range arrays {
		if writeKey {
			buffer.WriteString(fmt.Sprintf(`"%s" = `, keys[keyIdx]))
		}
		buffer.WriteString("ARRAY[")
		arrLen := len(arr)
		for arrIdx := range arr {
			buffer.WriteString(fmt.Sprintf("$%d", paramCount))
			if arrIdx+1 < arrLen {
				buffer.WriteString(", ")
			}
			paramCount++
		}

		//switch arr.(type)
		buffer.WriteString("]::text[]")
		if (idx + 1) < arrLength {
			buffer.WriteString(",\n")
		}
		idx++
	}
	buffer.WriteString("\n")
	return paramCount
}

func constructInsertQuery(e *eventhub.Event) string {

	var buffer bytes.Buffer
	buffer.WriteString(`
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
    `)
	writeArrayParams(
		&buffer,
		[][]string{
			e.Entities,
			e.OtherReferences,
			e.Actors,
			e.Tags,
		},
		6,
		nil)
	buffer.WriteString(`) RETURNING "id", "created", "updated";`)
	return buffer.String()
}

func constructUpdateQuery(e *eventhub.Event) string {

	var buffer bytes.Buffer
	buffer.WriteString(`UPDATE "event"
    SET
    "key" = $1,
    "payload" = $2,
    "description" = $3,
    "importance" = $4,
    "origin" = $5,`)

	nextParam := writeArrayParams(
		&buffer,
		[][]string{
			e.Entities,
			e.OtherReferences,
			e.Actors,
			e.Tags,
		},
		6,
		[]string{"entities", "other_references", "actors", "tags"})

	buffer.WriteString(fmt.Sprintf(`
        WHERE "id" = $%d
        RETURNING "updated";`, nextParam))
	return buffer.String()
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
	}

	for _, arr := range [][]string{e.Entities, e.OtherReferences, e.Actors, e.Tags} {
		for _, v := range arr {
			args = append(args, v)
		}
	}

	switch e.ID {
	case 0:
		err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
			return tx.QueryRow(constructInsertQuery(e), args...).Scan(&e.ID, &e.Created, &e.Updated)
		})
	default:
		args := append(args, e.ID)
		err = wrapTransaction(p.pg, func(tx *sql.Tx) error {
			return tx.QueryRow(constructUpdateQuery(e), args...).Scan(&e.Updated)
		})
	}

	return err
}
