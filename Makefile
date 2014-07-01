all: install

bindata.go: templates/tablestruct.go.tmpl
	go-bindata $(dir $<)

install: bindata.go
	go install

.PHONY: examples
examples:
	$(MAKE) -C examples/simple

.PHONY: clean
clean:
	rm -f bindata.go
