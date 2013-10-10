package db

import (
	"github.com/StefanKjartansson/eventhub"
	"testing"
)

type QueryBuilder func(e *eventhub.Event) (string, []interface{}, error)

func queryTest(t *testing.T, expected string, e *eventhub.Event, qb QueryBuilder) {

	expectedArgs := []interface{}{
		e.Key,
		"{}",
		"{}",
		e.Created,
		e.Updated,
		e.Description,
		e.Importance,
		e.Origin,
	}
	if e.ID > 0 {
		expectedArgs = []interface{}{
			e.Key,
			"{}",
			"{}",
			e.Description,
			e.Importance,
			e.Origin,
		}
	}

	for _, i := range [][]string{e.Entities, e.OtherReferences, e.Actors, e.Tags} {
		for _, s := range i {
			expectedArgs = append(expectedArgs, s)
		}

	}

	if e.ID > 0 {
		expectedArgs = append(expectedArgs, e.ID)
	}

	query, args, err := qb(e)

	if err != nil {
		t.Fatal(err)
	}

	//Need DeepEqual method for these cases
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %+v, got %+v", expectedArgs, args)
	}

	if query != expected {
		t.Fatalf("Expected:\n\t%s\nGot:\n\t%s", expected, query)
	}

}

func TestBuildInsertQuery(t *testing.T) {

	const expected = `insert into "event" ("key", "key_params", "created", "updated", "payload", "description", "importance", "origin", "entities", "other_references", "actors", "tags") values ($1, $2, $3, $4, $5, $6, $7, $8,ARRAY[$9, $10]::text[], ARRAY[]::text[], ARRAY[$11]::text[], ARRAY[]::text[]) returning "id", "created", "updated";`

	e := eventhub.NewEvent(
		"foo.bar",
		nil,
		nil,
		"My event",
		3,
		"mysystem",
		[]string{"ns/foo", "ns/moo"},
		nil,
		[]string{"someone"},
		nil)

	queryTest(t, expected, e, buildInsertQuery)

}

func TestBuildUpdateQuery(t *testing.T) {

	const expected = `update "event" set "key" = $1, "key_params" = $2, "payload" = $3, "description" = $4, "importance" = $5, "origin" = $6, "entities" = ARRAY[$7, $8]::text[], "other_references" = ARRAY[]::text[], "actors" = ARRAY[$9]::text[], "tags" = ARRAY[]::text[] where "id" = $10 returning "updated";`

	e := eventhub.NewEvent(
		"foo.bar",
		nil,
		nil,
		"My event",
		3,
		"mysystem",
		[]string{"ns/foo", "ns/moo"},
		nil,
		[]string{"someone"},
		nil)

	e.ID = 123

	queryTest(t, expected, e, buildUpdateQuery)

}
