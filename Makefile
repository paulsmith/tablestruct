all: install

install:
	go install ./...

check:
	go test -v

EXAMPLES = \
	examples/simple \
	examples/nullable \
	examples/multiple \
	examples/insertmany

examples: $(EXAMPLES)

.PHONY: examples $(EXAMPLES)
$(EXAMPLES):
	$(MAKE) -C $@
