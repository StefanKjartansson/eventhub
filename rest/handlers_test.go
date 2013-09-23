package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var serverAddr string
var once sync.Once
var client *http.Client

func startServer() {

	d := eventhub.NewDummyBackend()

	_ = d.Save(&eventhub.Event{
		Key:         "myapp.user.login",
		Description: "User foobar logged in",
		Importance:  3,
		Origin:      "myapp",
		Entities:    []string{"user/foo"},
	})

	_ = d.Save(&eventhub.Event{
		Key:         "myapp.user.logout",
		Description: "User foobar logged out",
		Importance:  3,
		Origin:      "myapp",
		Entities:    []string{"user/foo"},
	})

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

func TestGetByEntity(t *testing.T) {
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/user/foo/", serverAddr)
	events := []eventhub.Event{}
	getJSON(t, url, &events)
	log.Printf("%v", events)
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
