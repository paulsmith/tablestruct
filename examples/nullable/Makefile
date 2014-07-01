OBJECT_FILES = guestbook_mapper.go scanner.go

all: $(OBJECT_FILES) nullable

$(OBJECT_FILES): nullable.ts.json
	tablestruct $< ./

nullable: $(OBJECT_FILES) nullable.go
	go build

.PHONY: schema.sql
schema.sql:
	pg_dump -x -O -t person -s > $@

.PHONY: clean
clean:
	rm -f $(OBJECT_FILES) nullable
