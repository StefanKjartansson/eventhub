package eventhub

import (
	"log"
	"strconv"
	"strings"
	"unicode"
)

type MatchArray [2][]string

// Given two string arrays, returns true if the
// value array contains all of the query array's values
func (m MatchArray) Match() bool {

	qArr := m[0]
	eArr := m[1]

	//both are empty, ignore
	if len(qArr) == 0 && len(eArr) == 0 {
		return true
	}

	if len(qArr) > 0 {
		allMatch := true
		for _, s := range qArr {
			if !stringInSlice(s, eArr) {
				allMatch = false
			}
		}
		if !allMatch {
			return false
		}
	}
	return true

}

// Returns true if a string is in the slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Returns true if any of the OR delimited q string matches e:
//     orMatchStringAny("foo OR bar or baz", "baz")  // returns true
//     orMatchStringAny("foo", "foo")  // returns true
//     orMatchStringAny("  foo  ", "foo")  // returns true
//     orMatchStringAny("foo OR bar or baz", "moo")  // returns false
func orMatchStringAny(q, e string) bool {

	for _, s := range strings.Split(q, "OR") {
		s = strings.TrimSpace(s)
		if e == s {
			return true
		}
	}
	return false
}

//TODO: Created/Updated, lt & gt
//TODO: Importance, 1+OR+3 gt1 lt4. single param
//TODO: Sort & direction
type Query struct {
	Origin          string   `url:"origin,omitempty" schema:"origin" json:"origin"`
	Key             string   `url:"key,omitempty" schema:"key" json:"key"`
	Entities        []string `url:"entities,omitempty" schema:"entities" json:"entities"`
	OtherReferences []string `url:"other_references,omitempty" schema:"other_references" json:"other_references"`
	Actors          []string `url:"actors,omitempty" schema:"actors" json:"actors"`
	Tags            []string `url:"tags,omitempty" schema:"tags" json:"tags"`
	Importance      string   `url:"importance,omitempty" schema:"importance" json:"importance"`
}

// Returns true if the query values are empty
func (q *Query) IsEmpty() bool {

	for _, s := range []string{q.Origin, q.Key, q.Importance} {
		if s != "" {
			return false
		}
	}
	for _, arr := range [][]string{q.Entities, q.OtherReferences, q.Actors, q.Tags} {
		if len(arr) > 0 {
			return false
		}
	}
	return true

}

func (q *Query) matchImportance(i int) bool {

	if q.Importance == "" {
		return true
	}

	val, err := strconv.Atoi(strings.TrimFunc(q.Importance, unicode.IsLetter))
	if err != nil {
		//maybe don't fatal
		log.Fatal(err)
		return false
	}

	switch strings.TrimFunc(q.Importance, unicode.IsDigit) {
	case "gte":
		return (i >= val)
	case "gt":
		return (i > val)
	case "lte":
		return (i <= val)
	case "lt":
		return (i < val)
	}
	return (val == i)

}

// Determines whether an event matched the query
func (q *Query) Match(e Event) bool {

	// empty query matches everything
	if q.IsEmpty() {
		return true
	}

	orPairs := [][2]string{
		[2]string{q.Origin, e.Origin},
		[2]string{q.Key, e.Key},
	}

	for _, pair := range orPairs {

		//should orMatchStringAny return true if q (pair[0]) is empty?
		if pair[0] != "" && !orMatchStringAny(pair[0], pair[1]) {
			return false
		}

	}

	arrays := []MatchArray{
		{q.Entities, e.Entities},
		{q.OtherReferences, e.OtherReferences},
		{q.Actors, e.Actors},
		{q.Tags, e.Tags},
	}

	for _, ma := range arrays {
		if !ma.Match() {
			return false
		}
	}

	return q.matchImportance(e.Importance)
}
