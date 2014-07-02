all: install

install:
	go install

check:
	go test -v

.PHONY: examples
examples:
	$(MAKE) -C examples/simple

.PHONY: clean
clean:
	rm -f bindata.go
