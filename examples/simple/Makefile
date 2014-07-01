OBJECT_FILES = person_mapper.go scanner.go

all: $(OBJECT_FILES) simple

$(OBJECT_FILES): simple.ts.json
	tablestruct $< ./

simple: $(OBJECT_FILES) simple.go
	go build

.PHONY: schema.sql
schema.sql:
	pg_dump -x -O -t person -s > $@

.PHONY: clean
clean:
	rm -f $(OBJECT_FILES) simple
