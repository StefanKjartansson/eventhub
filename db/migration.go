package db

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Migration struct {
	filename string
	content  string
	date     time.Time
}

func (m Migration) String() string {
	return fmt.Sprintf("%s at %s", m.filename, m.date)
}

type Migrations []Migration

func (m Migrations) Len() int      { return len(m) }
func (m Migrations) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

//Returns a new Migration array with the provided dates filtered out
func (m Migrations) FilterDates(t []time.Time) (nm Migrations) {

	//Find the indices of the migrations which have already
	//been applied
	removalIndexes := []int{}
	for idx, im := range m {
		for _, it := range t {
			if im.date == it {
				removalIndexes = append(removalIndexes, idx)
			}
		}
	}

	nm = m[:]

	//Same length, return an empty set
	if len(removalIndexes) == len(nm) {
		return Migrations{}
	}

	//Swap & slice
	for _, idx := range removalIndexes {
		l := len(m) - 1
		nm.Swap(idx, l)
		nm = m[:l]
	}

	return nm
}

// Sort by date
type ByAge struct{ Migrations }

func (s ByAge) Less(i, j int) bool {
	return s.Migrations[i].date.Nanosecond() < s.Migrations[j].date.Nanosecond()
}

func globMigrations() (m Migrations, err error) {

	const longForm = "2006-01-02T15-04-05Z.sql"
	matches, err := filepath.Glob("./migrations/*.sql")

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	for _, s := range matches {

		datePart := strings.SplitAfterN(s, "-", 2)[1]
		t, _ := time.Parse(longForm, datePart)

		contents, err := ioutil.ReadFile(s)
		if err != nil {
			panic("unable to read a file")
		}

		x := Migration{s, string(contents), t}

		m = append(m, x)
	}

	sort.Sort(ByAge{m})

	return m, nil
}
