all: install

install:
	go install

.PHONY: examples
examples:
	$(MAKE) -C examples/simple

.PHONY: clean
clean:
	rm -f bindata.go
