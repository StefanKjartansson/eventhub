package rest

import (
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	"github.com/gorilla/mux"
	"log"
)

func GetRouter(d *eventhub.DataBackend) (*mux.Router, error) {

	r := mux.NewRouter()

	r.HandleFunc("/{entity}/{id}/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		vars := mux.Vars(req)
		filterParams := make(map[string]interface{})
		events, err := d.FilterBy(filterParams)
		if err != nil {
			log.Println(err)
		}
		enc := json.NewEncoder(w)
		enc.Encode(events)

	}).Methods("GET")

	return &r, nil
}
