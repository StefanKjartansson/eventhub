package eventhub

import "strings"

type MatchArray [2][]string

func (m MatchArray) Match(match bool) bool {

	qArr := m[0]
	eArr := m[1]

	if len(qArr) > 0 {
		allMatch := true
		for _, s := range qArr {
			if !stringInSlice(s, eArr) {
				allMatch = false
			}
		}
		match = allMatch
	}
	return match
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Query struct {
	Origin   string   `url:"origin,omitempty" schema:"origin" json:"origin"`
	Key      string   `url:"key,omitempty" schema:"key" json:"key"`
	Entities []string `url:"entities,omitempty" schema:"entities" json:"entities"`
	Actors   []string `url:"actors,omitempty" schema:"actors" json:"actors"`
}

// Determines whether an event matched the query
func (q *Query) Match(e Event) bool {

	//Catch non-initialized query fast
	if q.Origin == "" && q.Key == "" && len(q.Entities) == 0 && len(q.Actors) == 0 {
		return true
	}

	match := false
	if q.Origin != "" && e.Origin == q.Origin {
		match = true
	}

	if q.Key != "" {
		match = false
		for _, s := range strings.Split(q.Key, "OR") {
			s = strings.TrimSpace(s)
			if e.Key == s {
				match = true
			}
		}
		if match != true {
			return match
		}
	}

	arrays := []MatchArray{
		{q.Entities, e.Entities},
		{q.Actors, e.Actors},
	}

	for _, ma := range arrays {
		match = ma.Match(match)
	}

	return match
}
