all: test

test:
	gocov test github.com/StefanKjartansson/eventhub | gocov report
	gocov test github.com/StefanKjartansson/eventhub/db | gocov report
	gocov test github.com/StefanKjartansson/eventhub/rest | gocov report
	gocov test github.com/StefanKjartansson/eventhub/amqpfeed | gocov report
