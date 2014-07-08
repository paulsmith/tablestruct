package tablestruct

// ColumnMap describes a mapping between a Go struct field and a database
// column.
type ColumnMap struct {
	Field  string `json:"field"`
	Column string `json:"column"`
	Type   string `json:"type"`
	Null   bool   `json:"null"`
}
