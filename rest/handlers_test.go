package rest

import (
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	"github.com/gorilla/mux"
	"log"
)

var serverAddr string
var once sync.Once

func startServer() {
	router := GetRouter(&dummyBackend)
	http.Handle("/", router)
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test Server running on ", serverAddr)
}

func TestGetByEntity(t *testing.T) {
	once.Do(startServer)

}
