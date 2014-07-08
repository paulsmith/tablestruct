package tablestruct

import (
	"fmt"
	"strings"
)

// TableMap describes a mapping between a Go struct and database table.
type TableMap struct {
	Struct  string      `json:"struct"`
	Table   string      `json:"table"`
	Columns []ColumnMap `json:"columns"`
	// AutoPK is whether the table is set up to automatically generate new
	// values for the primary key column. `false` means the application must
	// supply them.
	AutoPK bool `json:"auto_pk"`
}

type importSpec struct {
	path string
	name string
}

func (i importSpec) String() string {
	if i.name == "" {
		return "\"" + i.path + "\""
	}
	return fmt.Sprintf("%s \"%s\"", i.name, i.path)
}

// ColumnList produces SQL for the column expressions in a SELECT statement.
func (t TableMap) ColumnList() string {
	cols := []string{t.PKCol()}
	for _, col := range t.Columns {
		cols = append(cols, col.Column)
	}
	return strings.Join(cols, ", ")
}

// UpdateList produces SQL for the column-placeholder pairs in a UPDATE
// statement.
func (t TableMap) UpdateList() string {
	cols := []string{t.PKCol() + " = $1"}
	for i, col := range t.Columns {
		cols = append(cols, fmt.Sprintf("%s = $%d", col.Column, i+2))
	}
	return strings.Join(cols, ", ")
}

// InsertList produces SQL for the placeholders in the value expression portion
// of an INSERT statement.
func (t TableMap) InsertList() string {
	var (
		vals   []string
		offset = 1
	)
	if t.AutoPK {
		vals = []string{"default"}
	} else {
		vals = []string{"$1"}
		offset = 2
	}
	for i := range t.Columns {
		vals = append(vals, fmt.Sprintf("$%d", i+offset))
	}
	return strings.Join(vals, ", ")
}

// InsertFields returns a list of struct fields to be used as values in an
// insert statement.
func (t TableMap) InsertFields() []string {
	var fields []string
	if !t.AutoPK {
		fields = []string{"ID"}
	}
	for i := range t.Columns {
		fields = append(fields, t.Columns[i].Field)
	}
	return fields
}

// PKCol returns the name of the primary key column.
func (t TableMap) PKCol() string {
	return "id"
}

// PKField returns the name of the primary key struct field.
func (t TableMap) PKField() string {
	return "ID"
}

// Fields returns the list of field names of the Go struct being mapped.
func (t TableMap) Fields() []string {
	f := []string{t.PKField()}
	for i := range t.Columns {
		f = append(f, t.Columns[i].Field)
	}
	return f
}
