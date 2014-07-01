OBJECT_FILES = person_mapper.go phone_mapper.go

$(OBJECT_FILES): simple.ts.json
	tablestruct $< ./

.PHONY: clean
clean:
	rm -f $(OBJECT_FILES)
