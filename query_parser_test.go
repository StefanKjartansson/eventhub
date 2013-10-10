package eventhub

import "testing"

func TestQueryParser(t *testing.T) {

	const input = "entities=c%2F1&entities=c%2F2&origin=mysystem&from=09.10.2013&to=09.10.2014"

	q, err := QueryFromString(input)

	t.Logf("%+v", *q)

	if err != nil {
		t.Fatal(err)
	}
}
