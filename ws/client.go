package ws

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"io"
)

var maxId int = 0

//to filter the feed
type FilterMessage struct {
	Entity string `json:"entity"`
}

func (fm *FilterMessage) Passes(event *eventhub.Event) bool {
	for _, e := range event.Entities {
		if fm.Entity == e {
			return true
		}
	}
	return false
}

//Represents a connected websocket client
type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	ch     chan *eventhub.Event
	doneCh chan bool
	filter FilterMessage
}

func NewClient(ws *websocket.Conn, server *Server) *Client {

	if ws == nil {
		panic("ws cannot be nil")
	}

	maxId++
	ch := make(chan *eventhub.Event)
	doneCh := make(chan bool)
	filter := FilterMessage{}

	return &Client{maxId, ws, server, ch, doneCh, filter}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(e *eventhub.Event) {
	select {
	case c.ch <- e:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.id)
		c.server.Err(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

func (c *Client) listenWrite() {

	for {
		select {

		case event := <-c.ch:
			err := websocket.JSON.Send(c.ws, event)
			if err != nil {
				c.server.Err(err)
			}

		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true
			return
		}
	}
}

func (c *Client) listenRead() {

	for {
		select {

		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true
			return

		// read data from websocket connection
		default:
			var fm FilterMessage
			err := websocket.JSON.Receive(c.ws, &fm)
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.server.Err(err)
			} else {
				c.filter = fm
			}
		}
	}
}
