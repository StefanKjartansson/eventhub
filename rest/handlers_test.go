package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
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

	d := eventhub.NewDummyBackend()

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

	_ = d.Save(&firstEvent)
	_ = d.Save(&secondEvent)

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

func getJSON(t *testing.T, url string, v interface{}) {

	t.Logf("[GET]: %s\n", url)

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

func postJSON(t *testing.T, url string, v interface{}) *http.Response {

	t.Logf("[POST]: %s\n", url)

	buf, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to serialize %v to json", v)
	}

	log.Printf("Posting JSON: %s", string(buf))

	r, err := client.Post(url, "application/json", bytes.NewReader(buf))

	if err != nil {
		t.Errorf("Error when posting to %s, error: %v", url, err)
	}

	return r
}

func putJSON(t *testing.T, url string, v interface{}) *http.Response {

	t.Logf("[PUT]: %s\n", url)

	buf, err := json.Marshal(v)
	if err != nil {
		t.Errorf("Unable to serialize %v to json", v)
	}

	log.Printf("PUT JSON: %s", string(buf))

	req, err := http.NewRequest("PUT", url, bytes.NewReader(buf))

	if err != nil {
		t.Errorf("PUT, Error to %s, error: %v", url, err)
	}

	r, err := client.Do(req)

	if err != nil {
		t.Errorf("Error when posting to %s, error: %v", url, err)
	}

	return r
}

func TestGetByEntity(t *testing.T) {
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/user/foo/", serverAddr)
	events := []eventhub.Event{}
	getJSON(t, url, &events)
	log.Printf("%v", events)
}

func TestGetById(t *testing.T) {
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/1/", serverAddr)
	event := eventhub.Event{}
	getJSON(t, url, &event)
	log.Printf("%v", event)
}

func TestPostNewEvent(t *testing.T) {
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
