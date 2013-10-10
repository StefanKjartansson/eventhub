package straumur

import (
	"github.com/gorilla/schema"
	"log"
	"net/url"
	"reflect"
	"time"
)

const dateForm = "02.01.2006"

func convertTime(value string) reflect.Value {

	//todo, more forms of dates?
	t, err := time.Parse(dateForm, value)

	if err != nil {
		log.Println(err)
	}

	return reflect.ValueOf(t)
}

func QueryFromValues(u url.Values) (*Query, error) {
	decoder := schema.NewDecoder()
	decoder.RegisterConverter(time.Time{}, convertTime)
	q := new(Query)
	err := decoder.Decode(q, u)
	return q, err
}

func QueryFromString(s string) (*Query, error) {
	u, err := url.ParseQuery(s)
	if err != nil {
		return nil, err
	}
	return QueryFromValues(u)
}
