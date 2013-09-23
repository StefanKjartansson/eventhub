package rest

import (
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

	router, err := GetRouter(d)

	if err != nil {
		panic(err)
	}

	http.Handle("/", router)
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test Server running on ", serverAddr)
}

func testRequest(t *testing.T, url string, v interface{}) {

	t.Logf("[GET]: %s\n", url)

	r, err := http.Get(url)

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

func TestGetByEntity(t *testing.T) {
	once.Do(startServer)

	url := fmt.Sprintf("http://%s/user/foo/", serverAddr)
	events := []eventhub.Event{}
	testRequest(t, url, &events)
	log.Printf("%v", events)
}
