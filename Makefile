all: test

test:
	go test -i github.com/straumur/straumur
	go test -i github.com/straumur/straumur/db
	go test -i github.com/straumur/straumur/rest
	go test -i github.com/straumur/straumur/amqpfeed
	go test -i github.com/straumur/straumur/ws
	go test github.com/straumur/straumur
	go test github.com/straumur/straumur/db
	go test github.com/straumur/straumur/rest
	go test github.com/straumur/straumur/amqpfeed
	go test github.com/straumur/straumur/ws
