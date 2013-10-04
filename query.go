package eventhub

import "strings"

type MatchArray [2][]string

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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func orMatchStringAny(q, e string) bool {

	for _, s := range strings.Split(q, "OR") {
		s = strings.TrimSpace(s)
		if e == s {
			return true
		}
	}
	return false
}

//TODO: plugin other_references & tags
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
	Updated         string   `url:"updated,omitempty" schema:"updated" json:"updated"`
	Created         string   `url:"created,omitempty" schema:"created" json:"created"`
}

func (q *Query) IsEmpty() bool {

	if q.Origin == "" && q.Key == "" && len(q.Entities) == 0 && len(q.Actors) == 0 {
		return true
	}
	return false
}

// Determines whether an event matched the query
func (q *Query) Match(e Event) bool {

	if q.IsEmpty() {
		return true
	}

	orPairs := [][2]string{
		[2]string{q.Origin, e.Origin},
		[2]string{q.Key, e.Key},
	}

	for _, pair := range orPairs {
		if pair[0] != "" && !orMatchStringAny(pair[0], pair[1]) {
			return false
		}
	}

	arrays := []MatchArray{
		{q.Entities, e.Entities},
		{q.Actors, e.Actors},
	}

	for _, ma := range arrays {
		if !ma.Match() {
			return false
		}
	}
	return true
}
