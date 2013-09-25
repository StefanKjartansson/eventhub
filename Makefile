all: test

test:
	go test -i github.com/StefanKjartansson/eventhub
	go test -i github.com/StefanKjartansson/eventhub/db
	go test -i github.com/StefanKjartansson/eventhub/rest
	go test -i github.com/StefanKjartansson/eventhub/amqpfeed
	go test -i github.com/StefanKjartansson/eventhub/ws
	go test github.com/StefanKjartansson/eventhub
	go test github.com/StefanKjartansson/eventhub/db
	go test github.com/StefanKjartansson/eventhub/rest
	go test github.com/StefanKjartansson/eventhub/amqpfeed
	go test github.com/StefanKjartansson/eventhub/ws
