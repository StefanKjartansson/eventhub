package db

import (
	"github.com/StefanKjartansson/eventhub"
	"testing"
)

func TestAggregateBuilder(t *testing.T) {

	const expected = `select i as name, count(*) as count from (select unnest(actors) as i from (select * from event where key in ($1, $2) and origin = $3 and entities @> ARRAY[$4, $5]::text[] order by updated desc) x order by actors) t group by i order by i;`

	expectedArgs := []interface{}{"foo.bar", "bar.foo", "mysystem", "c/1", "c/2"}

	q := eventhub.Query{}
	q.Origin = "mysystem"
	q.Entities = []string{"c/1", "c/2"}
	q.Key = "foo.bar OR bar.foo"

	s, args := buildAggregateQuery(q, "actors")

	if s != expected {
		t.Fatalf("Expected %s, got %s", expected, s)
	}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %+v, got %+v", expectedArgs, args)
	}

}
