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
	var cols []string
	for _, col := range t.Columns {
		cols = append(cols, col.Column)
	}
	return strings.Join(cols, ", ")
}

// UpdateList produces SQL for the column-placeholder pairs in a UPDATE
// statement.
func (t TableMap) UpdateList() string {
	var cols []string
	for i, col := range t.Columns {
		cols = append(cols, fmt.Sprintf("%s = $%d", col.Column, i+1))
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
	offset = 0
	for i := range t.Columns {
		if t.AutoPK && t.Columns[i].PrimaryKey {
			vals = append(vals, "default")
			continue
		}
		offset++
		vals = append(vals, fmt.Sprintf("$%d", offset))
	}
	return strings.Join(vals, ", ")
}

// InsertFields returns a list of struct fields to be used as values in an
// insert statement.
func (t TableMap) InsertFields() []string {
	var fields []string
	for i := range t.Columns {
		if t.AutoPK && t.Columns[i].PrimaryKey {
			continue
		}
		fields = append(fields, t.Columns[i].Field)
	}
	return fields
}

// Fields returns the list of field names of the Go struct being mapped.
func (t TableMap) Fields() []string {
	var f []string
	for i := range t.Columns {
		f = append(f, t.Columns[i].Field)
	}
	return f
}

// PrimaryKey returns the column mapping for the primary key field/column.
func (t TableMap) PrimaryKey() *ColumnMap {
	for i := range t.Columns {
		if t.Columns[i].PrimaryKey {
			return &t.Columns[i]
		}
	}
	return nil
}

// StructToTable converts a Go struct name to a database table name. It is
// mainly CamelCase -> snake_case, with some special cases, and is overridable.
func StructToTable(strct string) string {
	// TODO(paulsmith): implement what the doc comment actually says.
	return strings.ToLower(strct)
}
