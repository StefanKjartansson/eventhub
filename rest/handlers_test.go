package rest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"github.com/StefanKjartansson/eventhub/db"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

var serverAddr string
var once sync.Once
var client *http.Client
var firstEvent eventhub.Event
var secondEvent eventhub.Event

func startServer() {

	const connection = "dbname=teststream host=localhost sslmode=disable"
	d, err := db.NewPostgresDataSource(connection)

	//d := eventhub.NewDummyBackend()
	pdb, err := sql.Open("postgres", connection)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case _ = <-d.Updates():
			}
		}
	}()

	_, err = pdb.Exec(`TRUNCATE TABLE event RESTART IDENTITY;`)
	if err != nil {
		panic(err)
	}

	firstEvent = eventhub.Event{
		Key:         "myapp.user.login",
		Description: "User foobar logged in",
		Importance:  3,
		Origin:      "myapp",
		Entities:    []string{"user/foo"},
	}

	secondEvent = eventhub.Event{
		Key:         "myapp.user.logout",
		Description: "User foobar logged out",
		Importance:  2,
		Origin:      "myapp",
		Entities:    []string{"user/foo"},
	}

	err = d.Save(&firstEvent)
	if err != nil {
		panic(err)
	}

	err = d.Save(&secondEvent)
	if err != nil {
		panic(err)
	}

	rest := NewRESTService(d)
	router, err := rest.GetRouter()

	if err != nil {
		panic(err)
	}

	http.Handle("/", router)
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test Server running on ", serverAddr)
	client = http.DefaultClient
	log.Print("Test Client created")
}

func makeRequest(t *testing.T, method, url string, v interface{}) *http.Response {

	var r *http.Response

	t.Logf("[%s]: %s", method, url)

	switch method {
	case "POST", "PUT":
		buf, err := json.Marshal(v)
		if err != nil {
			t.Errorf("Unable to serialize %v to json", v)
		}

		log.Printf("[%s] JSON: %s", method, string(buf))
		req, err := http.NewRequest(method, url, bytes.NewReader(buf))
		req.Header.Add("Content-Type", "application/json")
		if err != nil {
			t.Errorf("[%s] %s, error: %v", method, url, err)
		}
		r, err = client.Do(req)
		if err != nil {
			t.Errorf("Error when posting to %s, error: %v", url, err)
		}
	default:
		r, err := client.Get(url)
		if err != nil {
			t.Errorf("Error: %v\n", err)
		}

		if r.StatusCode != http.StatusOK {
			t.Errorf("Wrong status code: %d\n", r.StatusCode)
		}

		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err = dec.Decode(v)
		if err != nil {
			t.Errorf("Error: %v\n", err)
		}
	}

	return r

}

func getJSON(t *testing.T, url string, v interface{}) *http.Response {
	return makeRequest(t, "GET", url, v)
}

func postJSON(t *testing.T, url string, v interface{}) *http.Response {
	return makeRequest(t, "POST", url, v)
}

func putJSON(t *testing.T, url string, v interface{}) *http.Response {
	return makeRequest(t, "PUT", url, v)
}

func TestGetByEntity(t *testing.T) {
	log.Println("TestGetByEntity")
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/user/foo/", serverAddr)
	t.Log(url)
	events := []eventhub.Event{}
	getJSON(t, url, &events)
	log.Printf("%v", events)
}

func TestGetById(t *testing.T) {
	log.Println("TestGetById")
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/1/", serverAddr)
	t.Log(url)
	event := eventhub.Event{}
	getJSON(t, url, &event)
	log.Printf("%v", event)
}

func TestPostNewEvent(t *testing.T) {
	log.Println("TestPostNewEvent")
	once.Do(startServer)

	e := eventhub.Event{
		Key:         "myapp.user.delete",
		Description: "User foobar deleted",
		Importance:  3,
		Origin:      "myapp",
		Entities:    []string{"user/foo"},
	}

	url := fmt.Sprintf("http://%s/", serverAddr)
	r := postJSON(t, url, &e)
	if r.StatusCode != http.StatusCreated {
		t.Errorf("Status code expected %d, got %d", http.StatusCreated, r.StatusCode)
	}
	log.Print("Post succeded")
}

func TestPutEvent(t *testing.T) {
	log.Println("TestPutEvent")
	once.Do(startServer)

	firstEvent.Description = "baz bar foo"
	url := fmt.Sprintf("http://%s/1/", serverAddr)
	r := putJSON(t, url, &firstEvent)
	if r.StatusCode != http.StatusAccepted {
		t.Errorf("Status code expected %d, got %d", http.StatusCreated, r.StatusCode)
	}
	log.Print("PUT succeded")
}

func TestSearch(t *testing.T) {
	log.Println("TestSearch")
	once.Do(startServer)

	tests := []struct {
		Params url.Values
		Status int
	}{{
		Params: url.Values{
			"key": {"myapp.user.login"},
		},
		Status: http.StatusOK,
	}, {
		Params: url.Values{
			"Key": {"myapp.user.login"},
		},
		Status: http.StatusOK,
	}, {
		Params: url.Values{
			"key": {"myapp.user.login", "myapp.user.logout"},
		},
		Status: http.StatusOK,
	}}

	for _, test := range tests {
		url := fmt.Sprintf("http://%s/search?%s", serverAddr, test.Params.Encode())
		results := []eventhub.Event{}
		getJSON(t, url, &results)
	}

	values := url.Values{
		"key": {"myapp.user.login", "myapp.user.logout"},
	}
	url := fmt.Sprintf("http://%s/user/foo/search?%s", serverAddr, values.Encode())
	results := []eventhub.Event{}
	getJSON(t, url, &results)

}
