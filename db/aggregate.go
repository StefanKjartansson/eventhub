package db

import (
	"bytes"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"strings"
)

func buildAggregateQuery(q eventhub.Query, arrayType string) (string, []interface{}) {

	var buffer bytes.Buffer
	buffer.WriteString("select i as name, count(*) as count from ")
	buffer.WriteString(fmt.Sprintf("(select unnest(%s) as i from (", arrayType))
	sel, args := buildSelectQuery(q)
	buffer.WriteString(strings.TrimRight(sel, ";"))
	buffer.WriteString(fmt.Sprintf(") x order by %s) t group by i order by i;", arrayType))
	return buffer.String(), args
}
