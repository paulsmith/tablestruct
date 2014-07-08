package tablestruct

import "strings"

// ColumnMap describes a mapping between a Go struct field and a database
// column.
type ColumnMap struct {
	Field      string `json:"field"`
	Column     string `json:"column"`
	Type       string `json:"type"`
	Null       bool   `json:"null"`
	PrimaryKey bool   `json:"pk"`
}

// FieldToColumn converts a Go struct field name to a database table column
// name. It is mainly CamelCase -> snake_case, with some special cases, and is
// overridable.
func FieldToColumn(field string) string {
	// TODO(paulsmith): implement what the doc comment actually says.
	return strings.ToLower(field)
}
