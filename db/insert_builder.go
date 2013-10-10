package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	_ "github.com/lib/pq"
	"time"
)

func writeArrayParams(buffer *bytes.Buffer, arrays [][]string, startParam int, keys []string) int {

	//todo, accept more than just string arrays
	//todo, remove panic

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
			buffer.WriteString(", ")
		}
		idx++
	}
	return paramCount
}

func getBaseArgs(e *eventhub.Event, isUpdate bool) ([]interface{}, error) {

	//todo, check if e.Payload/e.KeyParams is string and valid json
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, err
	}

	keyparams, err := json.Marshal(e.KeyParams)
	if err != nil {
		return nil, err
	}

	if isUpdate {

		return []interface{}{
			e.Key,
			keyparams,
			payload,
			e.Description,
			e.Importance,
			e.Origin,
		}, nil

	}
	t := time.Now()
	if e.Created.IsZero() {
		e.Created = t
	}
	e.Updated = t

	return []interface{}{
		e.Key,
		keyparams,
		e.Created,
		e.Updated,
		payload,
		e.Description,
		e.Importance,
		e.Origin,
	}, nil

}

// Builds an insert query for the event table
func buildInsertQuery(e *eventhub.Event) (string, []interface{}, error) {

	var buffer bytes.Buffer
	columnList := []string{
		"key",
		"key_params",
		"created",
		"updated",
		"payload",
		"description",
		"importance",
		"origin",
		"entities",
		"other_references",
		"actors",
		"tags",
	}

	args, err := getBaseArgs(e, false)
	if err != nil {
		return "", nil, nil
	}

	buffer.WriteString(`insert into "event" (`)
	for arrIdx, s := range columnList {
		buffer.WriteString(fmt.Sprintf(`"%s"`, s))
		if arrIdx+1 < len(columnList) {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(") values (")
	buffer.WriteString("$1, $2, $3, $4, $5, $6, $7, $8,")

	writeArrayParams(
		&buffer,
		[][]string{
			e.Entities,
			e.OtherReferences,
			e.Actors,
			e.Tags,
		},
		len(args)+1,
		nil)

	buffer.WriteString(`) returning "id", "created", "updated";`)

	for _, arr := range [][]string{e.Entities, e.OtherReferences, e.Actors, e.Tags} {
		for _, v := range arr {
			args = append(args, v)
		}
	}

	return buffer.String(), args, nil
}

// Builds an update query for the event table
func buildUpdateQuery(e *eventhub.Event) (string, []interface{}, error) {

	//Partial updates?

	var buffer bytes.Buffer
	paramCount := 1
	columnList := []string{
		"key",
		"key_params",
		"payload",
		"description",
		"importance",
		"origin",
	}

	args, err := getBaseArgs(e, true)

	if err != nil {
		return "", nil, nil
	}

	buffer.WriteString(`update "event" set `)
	for _, s := range columnList {
		buffer.WriteString(fmt.Sprintf(`"%s" = $%d, `, s, paramCount))
		paramCount++
	}

	nextParam := writeArrayParams(
		&buffer,
		[][]string{
			e.Entities,
			e.OtherReferences,
			e.Actors,
			e.Tags,
		},
		paramCount,
		[]string{"entities", "other_references", "actors", "tags"})

	for _, arr := range [][]string{e.Entities, e.OtherReferences, e.Actors, e.Tags} {
		for _, v := range arr {
			args = append(args, v)
		}
	}

	args = append(args, e.ID)

	buffer.WriteString(fmt.Sprintf(` where "id" = $%d returning "updated";`, nextParam))

	return buffer.String(), args, nil
}
