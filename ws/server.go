package ws

import (
	"code.google.com/p/go.net/websocket"
	"github.com/StefanKjartansson/eventhub"
	"log"
	"net/http"
)

type Server struct {
	pattern string
	feed    eventhub.EventFeed
	clients map[int]*Client
	addCh   chan *Client
	delCh   chan *Client
	doneCh  chan bool
	errCh   chan error
}

// Create a new Websocket broadcaster
func NewServer(pattern string, feed eventhub.EventFeed) *Server {
	clients := make(map[int]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &Server{
		pattern,
		feed,
		clients,
		addCh,
		delCh,
		doneCh,
		errCh,
	}
}

func (s *Server) Add(c *Client) {
	s.addCh <- c
}

func (s *Server) Del(c *Client) {
	s.delCh <- c
}

func (s *Server) Done() {
	s.doneCh <- true
}

func (s *Server) Err(err error) {
	s.errCh <- err
}

func (s *Server) sendAll(event *eventhub.Event) {
	for _, c := range s.clients {
		if c.filter.Entity == "" || c.filter.Passes(event) {
			c.Write(event)
		}
	}
}

func (s *Server) Listen() {

	// websocket handler
	onConnected := func(ws *websocket.Conn) {

		defer func() {
			err := ws.Close()
			if err != nil {
				s.errCh <- err
			}
		}()

		client := NewClient(ws, s)
		s.Add(client)
		client.Listen()
	}

	http.Handle(s.pattern, websocket.Handler(onConnected))

	for {
		select {

		// Add new a client
		case c := <-s.addCh:
			log.Println("Added new client")
			s.clients[c.id] = c
			log.Println("Now", len(s.clients), "clients connected.")

		// del a client
		case c := <-s.delCh:
			log.Println("Delete client")
			delete(s.clients, c.id)

		// consume event feed
		case event := <-s.feed.Updates():
			log.Println("Send all:", event)
			s.sendAll(event)

		case err := <-s.errCh:
			log.Println("Error:", err.Error())

		case <-s.doneCh:
			return
		}
	}
}
