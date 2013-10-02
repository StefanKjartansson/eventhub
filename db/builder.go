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

func buildQuery(q eventhub.Query) (string, []interface{}) {

	var buffer bytes.Buffer
	args := []interface{}{}
	paramCount := 1
	delimiter := " and "
	writeDelimiter := false

	buffer.WriteString("select * from event where ")

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

	if len(q.Entities) > 0 {
		if writeDelimiter {
			buffer.WriteString(delimiter)
		}
		nextParam, s := writeArray(paramCount, &args, "entities", q.Entities)
		paramCount = nextParam
		buffer.WriteString(s)
		writeDelimiter = true
	}

	if len(q.Actors) > 0 {
		if writeDelimiter {
			buffer.WriteString(delimiter)
		}
		nextParam, s := writeArray(paramCount, &args, "actors", q.Actors)
		paramCount = nextParam
		buffer.WriteString(s)
		writeDelimiter = true
	}

	buffer.WriteString(" order by updated desc;")

	return buffer.String(), args
}
