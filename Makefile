all: test

test:
	go test -i github.com/straumur/straumur
	go test -i github.com/straumur/straumur/ws
	go test github.com/straumur/straumur
	go test github.com/straumur/straumur/ws
