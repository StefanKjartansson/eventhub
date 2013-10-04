package rest

//TODO: Better use of errchan
//TODO: Dont save directly to database, be a feed

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

func (r *RESTService) postHandler(w http.ResponseWriter, req *http.Request) {

	e, err := r.parseEvent(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = r.databackend.Save(&e)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (r *RESTService) updateHandler(w http.ResponseWriter, req *http.Request) {

	e, err := r.parseEvent(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if e.ID == 0 {
		http.Error(w, "No ID", http.StatusBadRequest)
		return
	}
	err = r.databackend.Save(&e)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
	router.HandleFunc(r.prefix+"/{entity}/{id}/", r.entityHandler).Methods("GET")
	router.HandleFunc(r.prefix+"/{entity}/{id}/search", r.entitySearchHandler).Methods("GET")
	router.HandleFunc(r.prefix+"/", r.postHandler).Methods("POST")
	router.HandleFunc(r.prefix+"/{id}/", r.retrieveHandler).Methods("GET")
	router.HandleFunc(r.prefix+"/{id}/", r.updateHandler).Methods("PUT")
	router.HandleFunc(r.prefix+"/search", r.searchHandler).Methods("GET")
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

	http.Handle("/", router)

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
	//Investigate way to perform graceful shutdown
	return nil
}
