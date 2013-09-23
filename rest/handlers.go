package rest

import (
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

var databackend eventhub.DataBackend

func entityHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(r)
	entity := vars["entity"]
	id := vars["id"]
	filterParams := make(map[string]interface{})
	filterParams["Entities"] = []string{strings.Join([]string{entity, id}, "/")}
	events, err := databackend.FilterBy(filterParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(events)
}

func postHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var e eventhub.Event
	err := decoder.Decode(&e)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = databackend.Save(&e)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetRouter(d eventhub.DataBackend) (*mux.Router, error) {

	databackend = d
	r := mux.NewRouter()
	r.HandleFunc("/{entity}/{id}/", entityHandler).Methods("GET")
	//r.HandleFunc("/{entity}/{id}/search", entitySearchHandler).Methods("GET")
	r.HandleFunc("/", postHandler).Methods("POST")
	//r.HandleFunc("/{id}/", retrieveHandler).Methods("GET")
	//r.HandleFunc("/{id}/", updateHandler).Methods("PUT")
	//r.HandleFunc("/search", searchHandler).Methods("GET")
	return r, nil
}
