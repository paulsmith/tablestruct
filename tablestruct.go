package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

// Map describes a mapping between database tables and Go structs.
type Map []TableMap

// Imports generates list of import specs required by generated code.
func (m Map) Imports() []importSpec {
	return []importSpec{
		{"database/sql", ""},
		{"github.com/lib/pq", "_"},
	}
}

// TableMap describes a mapping between a Go struct and database table.
type TableMap struct {
	Struct  string      `json:"struct"`
	Table   string      `json:"table"`
	Columns []ColumnMap `json:"columns"`
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
	cols := []string{}
	for i, col := range t.Columns {
		cols = append(cols, fmt.Sprintf("%s = $%d", col.Column, i+2))
	}
	return strings.Join(cols, ", ")
}

// InsertList produces SQL for the placeholders in the value expression portion
// of an INSERT statement.
func (t TableMap) InsertList() string {
	cols := []string{"$1"}
	for i := range t.Columns {
		cols = append(cols, fmt.Sprintf("$%d", i+2))
	}
	return strings.Join(cols, ", ")
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
	dest string
}

func (c Code) write(format string, param ...interface{}) {
	c.buf.WriteString(fmt.Sprintf(format, param...))
}

type tableMapTmpl struct {
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

// Gen generates Go code for a set of table mappings and output them to files.
func (c Code) Gen(mapper Map, pkg string, filename string) {
	data := struct {
		Package   string
		Imports   []importSpec
		TableMaps []tableMapTmpl
	}{
		Package: pkg,
		Imports: mapper.Imports(),
	}

	for i, tableMap := range mapper {
		log.Printf("%d: generating map %s -> %s", i, tableMap.Table, tableMap.Struct)
		data.TableMaps = append(data.TableMaps, c.genMapper(tableMap))
	}

	if err := c.tmpl.Execute(c.buf, data); err != nil {
		log.Fatal(err)
	}

	path := c.dest + "/" + filename
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// gofmt
	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, "", c.buf.Bytes(), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	err = format.Node(f, fset, ast)
	if err != nil {
		log.Fatal(err)
	}
}

func (c Code) genMapper(mapper TableMap) tableMapTmpl {
	mapperFields := []string{"db *sql.DB"}
	return tableMapTmpl{
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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <metadata_file> <output_dir>\n", os.Args[0])
}

func main() {
	var (
		pkg      = flag.String("pkg", "main", "package of generated code")
		filename = flag.String("filename", "mapper.go", "name of generated mapper file")
	)

	flag.Usage = usage

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var mapper Map
	if err := json.NewDecoder(f).Decode(&mapper); err != nil {
		log.Fatal(err)
	}

	code := Code{
		buf:  bytes.NewBuffer(nil),
		tmpl: template.Must(template.New("tablestruct").Parse(mapperTemplate)),
		dest: flag.Arg(1),
	}
	code.Gen(mapper, *pkg, *filename)
}
