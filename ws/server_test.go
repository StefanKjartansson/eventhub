package ws

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"io"
	"log"
	"net/http/httptest"
	"sync"
	"syscall"
	"testing"
)

var serverAddr string
var once sync.Once

var d DummyFeed

type DummyFeed struct {
	events chan eventhub.Event
}

func (d *DummyFeed) Updates() <-chan eventhub.Event {
	return d.events
}

func (d *DummyFeed) Close() error {
	return nil
}

func startServer() {
	d = DummyFeed{make(chan eventhub.Event)}
	s := NewServer("/ws", &d)
	go s.Listen()
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
}

func TestWebSocketBroadcaster(t *testing.T) {

	once.Do(startServer)

	url := fmt.Sprintf("ws://%s%s", echoServerAddr, "/ws")
	conn, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		t.Errorf("WebSocket handshake error: %v", err)
		return
	}

	fm := FilterMessage{"ns/moo"}
	websocket.JSON.Send(conn, fm)

	e := eventhub.Event{
		Key:         "foo.bar",
		Description: "ba ba",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
	}
	d.events <- e

	var event eventhub.Event
	if err := websocket.JSON.Receive(conn, &event); err != nil {
		t.Errorf("Read: %v", err)
	}

	incoming := make(chan eventhub.Event)
	go readEvents(conn, incoming)

	d.events <- eventhub.Event{
		Key:         "Should filter",
		Description: "ba ba",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/boo"},
	}

	d.events <- eventhub.Event{
		Key:         "foo.bar",
		Description: "ba ba",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
	}

	ev := <-incoming

	if ev.Key != "foo.bar" {
		t.Errorf("Unexpected %s", ev)
	}

}

func readEvents(ws *websocket.Conn, incoming chan eventhub.Event) {
	for {
		var event eventhub.Event
		err := websocket.JSON.Receive(ws, &event)
		if err == nil {
			log.Println(event)
			incoming <- event
			continue
		}
		if err == io.EOF || err == syscall.EINVAL || err == syscall.ECONNRESET {
			log.Println("Peer disconnected", err.Error())
			return
		}
	}
}
