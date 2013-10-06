package db

import (
	"bytes"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"strings"
)

func writeArray(paramCount int, args *[]interface{}, key string, arr []string) (int, string) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s @> ARRAY[", key))
	for arrIdx, i := range arr {
		buffer.WriteString(fmt.Sprintf("$%d", paramCount))
		if arrIdx+1 < len(arr) {
			buffer.WriteString(", ")
		}
		*args = append(*args, i)
		paramCount++
	}
	buffer.WriteString("]::text[]")
	return paramCount, buffer.String()
}

func buildSelectQuery(q eventhub.Query) (string, []interface{}) {

	var buffer bytes.Buffer
	args := []interface{}{}
	paramCount := 1
	delimiter := " and "
	writeDelimiter := false

	buffer.WriteString("select * from event")

	//Early exit, query is empty
	if q.IsEmpty() {
		buffer.WriteString(" order by updated desc;")
		return buffer.String(), args
	}

	buffer.WriteString(" where ")

	if q.Key != "" {
		buffer.WriteString("key in (")
		keys := strings.Split(q.Key, "OR")
		for arrIdx, s := range keys {
			args = append(args, strings.TrimSpace(s))
			buffer.WriteString(fmt.Sprintf("$%d", paramCount))
			paramCount++
			if arrIdx+1 < len(keys) {
				buffer.WriteString(", ")
			}
		}
		buffer.WriteString(")")
		writeDelimiter = true
	}

	if q.Origin != "" {
		if writeDelimiter {
			buffer.WriteString(delimiter)
		}
		buffer.WriteString(fmt.Sprintf("origin = $%d", paramCount))
		args = append(args, q.Origin)
		paramCount++
		writeDelimiter = true
	}

	//Array fields
	array_fields := make(map[string][]string)
	array_fields["entities"] = q.Entities
	array_fields["actors"] = q.Actors

	for db_name, array_field := range array_fields {

		if len(array_field) > 0 {
			if writeDelimiter {
				buffer.WriteString(delimiter)
			}
			nextParam, s := writeArray(paramCount, &args, db_name, array_field)
			paramCount = nextParam
			buffer.WriteString(s)
			writeDelimiter = true
		}

	}

	//todo add sort to query
	buffer.WriteString(" order by updated desc;")

	return buffer.String(), args
}
