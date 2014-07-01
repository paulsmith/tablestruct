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

// Imports generates list of import specs required by generated code.
func (t TableMap) Imports() []importSpec {
	return []importSpec{
		{"database/sql", ""},
		{"github.com/lib/pq", "_"},
	}
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

// Fields returns the list of field names of the Go struct being mapped.
func (t TableMap) Fields() []string {
	f := []string{"ID"}
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

// Gen generates Go code for a set of table mappings and output them to files.
func (c Code) Gen(mapper Map, pkg string) {
	for i, tableMap := range mapper {
		log.Printf("%d: generating %s", i, tableMap.Struct)
		c.genMapper(tableMap, pkg)

		filename := strings.ToLower(tableMap.Struct) + "_mapper.go"
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

		c.buf.Reset()
	}

	// Auxiliary support for mapper - Scanner interface
	path := c.dest + "/" + "scanner.go"
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintf(f, "package %s\n\n", pkg)
	fmt.Fprintf(f, "type Scanner interface {\n")
	fmt.Fprintf(f, "\tScan(...interface{}) error\n")
	fmt.Fprintf(f, "}\n")
}

func (c Code) genMapper(mapper TableMap, pkg string) {
	mapperFields := []string{"db *sql.DB"}
	data := struct {
		Package      string
		Imports      []importSpec
		MapperType   string
		MapperFields []string
		VarName      string
		StructType   string
		ColumnList   string
		Table        string
		PKCol        string
		Fields       []string
		UpdateList   string
		UpdateCount  int
		InsertList   string
	}{
		pkg,
		mapper.Imports(),
		mapper.Struct + "Mapper",
		mapperFields,
		strings.ToLower(mapper.Struct[0:1]),
		mapper.Struct,
		mapper.ColumnList(),
		mapper.Table,
		mapper.PKCol(),
		mapper.Fields(),
		mapper.UpdateList(),
		len(mapper.Columns) + 1,
		mapper.InsertList(),
	}
	if err := c.tmpl.Execute(c.buf, data); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <metadata_file> <output_dir>\n", os.Args[0])
}

func main() {
	var (
		pkg = flag.String("pkg", "main", "package of generated code")
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
	code.Gen(mapper, *pkg)
}
