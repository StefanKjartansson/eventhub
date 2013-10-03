package db

import (
	"github.com/StefanKjartansson/eventhub"
	"testing"
)

func TestWriteArray(t *testing.T) {
	const expected = "foo @> ARRAY[$1, $2, $3]::text[]"
	args := []interface{}{}
	nextParam, s := writeArray(1, &args, "foo", []string{"a", "b", "c"})
	if s != expected {
		t.Fatalf("Expected %s, got %s", expected, s)
	}
	if nextParam != 4 {
		t.Fatalf("Expected nextParam to be 4, got %d", nextParam)
	}
}

func TestQueryBuilder(t *testing.T) {

	const expected = `select * from event where key in ($1, $2) and origin = $3 and entities @> ARRAY[$4, $5]::text[] order by updated desc;`

	expectedArgs := []interface{}{"foo.bar", "bar.foo", "mysystem", "c/1", "c/2"}

	q := eventhub.Query{}
	q.Origin = "mysystem"
	q.Entities = []string{"c/1", "c/2"}
	q.Key = "foo.bar OR bar.foo"

	s, args := buildSelectQuery(q)

	if s != expected {
		t.Fatalf("Expected %s, got %s", expected, s)
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %+v, got %+v", expectedArgs, args)
	}
}
