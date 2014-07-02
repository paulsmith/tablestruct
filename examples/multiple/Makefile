OBJECT_FILES = mapper.go

all: $(OBJECT_FILES) multiple

$(OBJECT_FILES): multiple.ts.json
	tablestruct < $< > $@

multiple: $(OBJECT_FILES) multiple.go
	go build

.PHONY: schema.sql
schema.sql:
	pg_dump -x -O -t person -t phone -s > $@

.PHONY: clean
clean:
	rm -f $(OBJECT_FILES) multiple
