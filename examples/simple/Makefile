OBJECT_FILES = person_mapper.go phone_mapper.go scanner.go

all: $(OBJECT_FILES) simple

$(OBJECT_FILES): simple.ts.json
	tablestruct $< ./

simple: $(OBJECT_FILES)
	go build

.PHONY: clean
clean:
	rm -f $(OBJECT_FILES) simple
