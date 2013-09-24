package ws

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"fmt"
	"github.com/StefanKjartansson/eventhub"
	"log"
	"net"
	"net/http/httptest"
	"sync"
	"testing"
)

var serverAddr string
var once sync.Once
var d DummyFeed

func NewWebsocketClient(t *testing.T) *websocket.Conn {
	client, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	config, err := websocket.NewConfig(fmt.Sprintf("websocket://%s%s", serverAddr, "/ws"), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := websocket.NewClient(config, client)
	if err != nil {
		t.Fatalf("WebSocket handshake error: %v", err)
	}
	return conn
}

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
	s := NewWebsocketServer("/ws", &d)
	go s.Listen()
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test Server running on ", serverAddr)
}

func TestWebSocketBroadcaster(t *testing.T) {

	once.Do(startServer)
	wsclient := NewWebsocketClient(t)

	e := eventhub.Event{
		Key:         "foo.bar",
		Description: "ba ba",
		Importance:  3,
		Origin:      "mysystem",
		Entities:    []string{"ns/foo", "ns/moo"},
	}

	b, err := json.Marshal(e)

	if err != nil {
		t.Fatal(err)
	}

	d.events <- e

	var actual_msg = make([]byte, len(b))
	n, err := wsclient.Read(actual_msg)
	t.Log(n)
	if err != nil {
		t.Errorf("Read: %v", err)
	}

	t.Fatalf("Expected %s, got %s", string(b), string(actual_msg))
}
