all: install

install:
	go install ./...

check:
	go test -v
