package ws

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"log"
	"net/http"
)

var maxId int = 0

type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	ch     chan *eventhub.Event
	doneCh chan bool
}

func NewClient(ws *websocket.Conn, server *Server) *Client {

	maxId++
	ch := make(chan *eventhub.Event)
	doneCh := make(chan bool)
	return &Client{maxId, ws, server, ch, doneCh}
}

func (c *Client) listenWrite() {
	log.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.ch:
			log.Println("Send:", msg)
			websocket.JSON.Send(c.ws, msg)

		// receive done request
		case <-c.doneCh:
			log.Println("Send:")

			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

func (c *Client) Listen() {
	go c.listenWrite()
}

func (c *Client) Write(msg *eventhub.Event) {
	log.Println("For client ", msg)
	websocket.JSON.Send(c.ws, msg)
	c.ch <- msg
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.id)
		c.server.Err(err)
	}
}

type Server struct {
	pattern string
	feed    eventhub.EventFeed
	clients map[int]*Client
	addCh   chan *Client
	delCh   chan *Client
	doneCh  chan bool
	errCh   chan error
}

// Create new chat server.
func NewWebsocketServer(pattern string, feed eventhub.EventFeed) *Server {
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

func (s *Server) Err(err error) {
	s.errCh <- err
}

func (s *Server) sendAll(e *eventhub.Event) {
	for _, c := range s.clients {
		c.Write(e)
	}
}

func (s *Server) Listen() {

	log.Println("Listening server...")

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
	log.Println("Created handler")

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

		case err := <-s.errCh:
			log.Println("Error:", err.Error())

		case e := <-s.feed.Updates():
			log.Println("Got event")
			s.sendAll(&e)

		case <-s.doneCh:
			return
		}
	}
}
