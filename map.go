package tablestruct

import (
	"encoding/json"
	"io"
)

// Map describes a mapping between database tables and Go structs.
type Map []TableMap

// NewMap constructs a new mapping object.
func NewMap(in io.Reader) (*Map, error) {
	var mapper Map
	if err := json.NewDecoder(in).Decode(&mapper); err != nil {
		return nil, err
	}
	return &mapper, nil
}

// Imports generates list of import specs required by generated code.
func (m *Map) Imports() []importSpec {
	return []importSpec{
		{"database/sql", ""},
		{"log", ""},
		{"github.com/lib/pq", "_"},
	}
}
