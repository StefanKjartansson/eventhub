package db

import (
	"database/sql"
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

func bootstrapDatabase(db *sql.DB) {

	// Get all table names
	// TODO: maybe change the schema name?

	rows, err := db.Query(`
        select tablename
            from pg_tables
        where
            pg_tables.schemaname = 'public';
    `)

	if err != nil {
		log.Fatal(err)
	}

	canMigrate := false
	var s string
	for rows.Next() {
		rows.Scan(&s)
		if s == "migration_info" {
			canMigrate = true
		}
	}

	//No table names returned
	if s == "" {
		canMigrate = true
	}

	//Get the list of migrations
	m, err := globMigrations()

	if err != nil {
		log.Fatal(err)
	}

	//If there were tables, the migration_info
	//table should be among them
	if s != "" {
		rows, err := db.Query(`
            select created from
                migration_info
            order by created
        `)

		removalDates := []time.Time{}
		for rows.Next() {
			var t time.Time
			err = rows.Scan(&t)
			if err != nil {
				log.Fatal(err)
			}
			//Weird, table created with TZ, but Scan doesn't
			//add the UTC info
			removalDates = append(removalDates, t.UTC())
		}

		//Filter out migrations which have already been applied
		m = m.FilterDates(removalDates)
	}

	//Run migrations
	if canMigrate && len(m) > 0 {

		for _, migration := range m {

			_, err := db.Exec(migration.content)
			if err != nil {
				log.Fatal(err)
			}
			_, err = db.Exec(`
                insert into migration_info
                    (created, content)
                values($1, $2)`, migration.date, migration.content)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
