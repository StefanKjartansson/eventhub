package rest

//TODO: Better use of errchan

import (
	"encoding/json"
	"github.com/StefanKjartansson/eventhub"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type RESTService struct {
	databackend eventhub.DataBackend
	address     string
	prefix      string
	events      chan *eventhub.Event
}

func (r *RESTService) entityHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(req)
	entity := vars["entity"]
	id := vars["id"]
	q := eventhub.Query{}
	q.Entities = append(q.Entities, strings.Join([]string{entity, id}, "/"))
	events, err := r.databackend.Query(q)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(events)

}

func (r *RESTService) retrieveHandler(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	vars := mux.Vars(req)
	id := vars["id"]
	idAsInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event, err := r.databackend.GetById(idAsInt)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(event)

}

func (r *RESTService) parseEvent(body io.ReadCloser) (eventhub.Event, error) {
	decoder := json.NewDecoder(body)
	defer body.Close()
	var e eventhub.Event
	err := decoder.Decode(&e)
	return e, err
}

func (r *RESTService) saveHandler(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	id := vars["id"]

	e, err := r.parseEvent(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id == "" && e.ID != 0 {
		http.Error(w, "Can't POST existing resource", http.StatusBadRequest)
		return
	}

	if id != "" && e.ID == 0 {
		http.Error(w, "No ID for update", http.StatusBadRequest)
		return
	}

	switch e.ID {
	case 0:
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusAccepted)
	}

	r.events <- &e
}

func (r *RESTService) search(q eventhub.Query, entity string) ([]*eventhub.Event, error) {
	if entity != "" {
		q.Entities = append(q.Entities, entity)
	}
	return r.databackend.Query(q)
}

func (r *RESTService) searchHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	q := new(eventhub.Query)
	decoder := schema.NewDecoder()
	err := decoder.Decode(q, req.Form)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events, err := r.search(*q, "")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		http.NotFound(w, req)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(events)
}

func (r *RESTService) entitySearchHandler(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	vars := mux.Vars(req)
	entity := vars["entity"]
	id := vars["id"]

	q := new(eventhub.Query)
	decoder := schema.NewDecoder()
	err := decoder.Decode(q, req.Form)

	events, err := r.search(*q, strings.Join([]string{entity, id}, "/"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(events) == 0 {
		http.NotFound(w, req)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(events)
}

func (r *RESTService) getRouter() (*mux.Router, error) {

	router := mux.NewRouter()
	s := router.PathPrefix(r.prefix).Subrouter()
	s.HandleFunc("/{entity}/{id}/", r.entityHandler).Methods("GET")
	s.HandleFunc("/{entity}/{id}/search", r.entitySearchHandler).Methods("GET")
	s.HandleFunc("/", r.saveHandler).Methods("POST")
	s.HandleFunc("/{id}/", r.retrieveHandler).Methods("GET")
	s.HandleFunc("/{id}/", r.saveHandler).Methods("PUT")
	s.HandleFunc("/search", r.searchHandler).Methods("GET")
	return router, nil
}

func NewRESTService(prefix string, address string) *RESTService {
	return &RESTService{
		prefix:  prefix,
		address: address,
		events:  make(chan *eventhub.Event),
	}
}

//DataService interface
func (r *RESTService) Run(d eventhub.DataBackend, ec chan error) {

	r.databackend = d

	router, err := r.getRouter()
	if err != nil {
		ec <- err
	}

	http.Handle(r.prefix+"/", router)

	err = http.ListenAndServe(r.address, nil)
	if err != nil {
		ec <- err
	}

}

//EventFeed interface
func (r *RESTService) Updates() <-chan *eventhub.Event {
	return r.events
}

func (r *RESTService) Close() error {
	close(r.events)
	//Investigate way to perform graceful shutdown of http
	return nil
}
