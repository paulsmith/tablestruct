package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
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

// ColumnMap describes a mapping between a Go struct field and a database
// column.
type ColumnMap struct {
	Field  string `json:"field"`
	Column string `json:"column"`
	Type   string `json:"type"`
	Null   bool   `json:"null"`
}

// Code generates Go code that maps database tables to structs.
type Code struct {
	buf  *bytes.Buffer
	tmpl *template.Template
}

// NewCode creates a new code generator.
func NewCode() *Code {
	return &Code{
		buf:  bytes.NewBuffer(nil),
		tmpl: template.Must(template.New("tablestruct").Parse(mapperTemplate)),
	}
}

func (c *Code) write(format string, param ...interface{}) {
	c.buf.WriteString(fmt.Sprintf(format, param...))
}

type tableMapTmpl struct {
	Mapper       TableMap
	MapperType   string
	MapperFields []string
	VarName      string
	StructType   string
	ColumnList   string
	Table        string
	PKCol        string
	PKField      string
	Fields       []string
	UpdateList   string
	UpdateCount  int
	InsertList   string
}

// Gen generates Go code for a set of table mappings.
func (c *Code) Gen(mapper *Map, pkg string, out io.Writer) {
	data := struct {
		Package   string
		Imports   []importSpec
		TableMaps []tableMapTmpl
	}{
		Package: pkg,
		Imports: mapper.Imports(),
	}

	for i, tableMap := range *mapper {
		log.Printf("%d: generating map %s -> %s", i, tableMap.Table, tableMap.Struct)
		data.TableMaps = append(data.TableMaps, c.genMapper(tableMap))
	}

	if err := c.tmpl.Execute(c.buf, data); err != nil {
		// TODO(paulsmith): return error
		log.Fatal(err)
	}

	// gofmt
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, "", c.buf.Bytes(), parser.ParseComments)
	if err != nil {
		// TODO(paulsmith): return error
		log.Fatal(err)
	}
	err = format.Node(out, fset, ast)
	if err != nil {
		// TODO(paulsmith): return error
		log.Fatal(err)
	}
}

func (c *Code) genMapper(mapper TableMap) tableMapTmpl {
	// TODO(paulsmith): move this.
	mapperFields := []string{
		"db *sql.DB",
		"sql map[string]string",
		"stmt map[string]*sql.Stmt",
	}
	return tableMapTmpl{
		mapper,
		mapper.Struct + "Mapper",
		mapperFields,
		strings.ToLower(mapper.Struct[0:1]),
		mapper.Struct,
		mapper.ColumnList(),
		mapper.Table,
		mapper.PKCol(),
		mapper.PKField(),
		mapper.Fields(),
		mapper.UpdateList(),
		len(mapper.Columns) + 1,
		mapper.InsertList(),
	}
}

func main() {
	var (
		pkg = flag.String("pkg", "main", "package of generated code")
	)

	flag.Parse()

	mapper, err := NewMap(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	code := NewCode()
	code.Gen(mapper, *pkg, os.Stdout)
}
