package eventstream

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"regexp"
	"strings"
)

const (
	insertStatement = `
    INSERT INTO 
        "event"
        ("key", "created", "payload", "description", "importance", "origin",
         "entities", "other_references", "actors", "tags")
    VALUES 
        ($1, $2, $3, $4, $5, $6, ARRAY[$7], ARRAY[$8], ARRAY[$9], ARRAY[$10]) 
    RETURNING "id";
    `

	selectByIdx = `
    SELECT "id", "key" 
    FROM
        "event"
    WHERE "id" = $1
    `
	selectById = `
    SELECT "id", "key", "created", "payload", "description", "importance", 
        "origin", "entities", "other_references", "actors", "tags"
    FROM
        "event"
    WHERE "id" = $1
    `
)

var (
	// unquoted array values must not contain: (" , \ { } whitespace NULL)
	// and must be at least one char
	unquotedChar  = `[^",\\{}\s(NULL)]`
	unquotedValue = fmt.Sprintf("(%s)+", unquotedChar)

	// quoted array values are surrounded by double quotes, can be any
	// character except " or \, which must be backslash escaped:
	quotedChar  = `[^"\\]|\\"|\\\\`
	quotedValue = fmt.Sprintf("\"(%s)*\"", quotedChar)

	// an array value may be either quoted or unquoted:
	arrayValue = fmt.Sprintf("(?P<value>(%s|%s))", unquotedValue, quotedValue)

	// Array values are separated with a comma IF there is more than one value:
	arrayExp = regexp.MustCompile(fmt.Sprintf("((%s)(,)?)", arrayValue))

	valueIndex int
)

type StringSlice []string

// Implements sql.Scanner for the String slice type
// Scanners take the database value (in this case as a byte slice)
// and sets the value of the type.  Here we cast to a string and
// do a regexp based parse
func (s *StringSlice) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return error(errors.New("Scan source was not []bytes"))
	}
	asString := string(asBytes)
	parsed := parseArray(asString)
	(*s) = StringSlice(parsed)
	return nil
}

func parseArray(array string) []string {
	results := make([]string, 0)
	matches := arrayExp.FindAllStringSubmatch(array, -1)
	for _, match := range matches {
		s := match[valueIndex]
		// the string _might_ be wrapped in quotes, so trim them:
		s = strings.Trim(s, "\"")
		results = append(results, s)
	}
	return results
}

type DataSource interface {
	Insert(e *Event) error
	GetById(id int) (*Event, error)
	FilterByKey(key string) (*[]Event, error)
}

type PostgresDataSource struct {
	pg         *sql.DB
	insert     *sql.Stmt
	selectbyid *sql.Stmt
}

func (p *PostgresDataSource) GetById(id int) (*Event, error) {

	var e Event
	var err error
	var tx *sql.Tx

	if tx, err = p.pg.Begin(); err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()
		}
	}()

	var entities StringSlice
	var references StringSlice
	var actors StringSlice
	var tags StringSlice

	err = tx.Stmt(p.selectbyid).QueryRow(id).Scan(
		&e.ID,
		&e.Key,
		&e.Created,
		&e.Payload,
		&e.Description,
		&e.Importance,
		&e.Origin,
		&entities,
		&references,
		&actors,
		&tags)

	if err != nil {
		return nil, err
	}

	e.Entities = entities
	e.OtherReferences = references
	e.Actors = actors
	e.Tags = tags

	return &e, nil

}

func (p *PostgresDataSource) Insert(e *Event) (err error) {

	var tx *sql.Tx

	if tx, err = p.pg.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()
		}
	}()

	b, err := json.Marshal(e.Payload)

	err = tx.Stmt(p.insert).QueryRow(
		e.Key,
		e.Created,
		b,
		e.Description,
		e.Importance,
		e.Origin,
		strings.Join(e.Entities, ", "),
		strings.Join(e.OtherReferences, ", "),
		strings.Join(e.Actors, ", "),
		strings.Join(e.Tags, ", ")).Scan(&e.ID)
	if err != nil {
		return err
	}

	return nil
}

func NewPostgresDataSource(connection string) (*PostgresDataSource, error) {

	p := PostgresDataSource{}

	pg, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	p.pg = pg

	bootstrapDatabase(p.pg)

	s, err := pg.Prepare(insertStatement)
	if err != nil {
		return &p, err
	}
	p.insert = s

	s, err = pg.Prepare(selectById)
	if err != nil {
		return &p, err
	}
	p.selectbyid = s

	return &p, nil
}
