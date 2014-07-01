package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
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

func (t TableMap) Imports() []importSpec {
	return []importSpec{
		{"database/sql", ""},
		{"github.com/lib/pq", "_"},
	}
}

func (t TableMap) ColumnList() string {
	cols := []string{t.PKCol()}
	for _, col := range t.Columns {
		cols = append(cols, col.Column)
	}
	return strings.Join(cols, ", ")
}

func (t TableMap) UpdateList() string {
	cols := []string{}
	for i, col := range t.Columns {
		cols = append(cols, fmt.Sprintf("%s = $%d", col.Column, i+2))
	}
	return strings.Join(cols, ", ")
}

func (t TableMap) InsertList() string {
	cols := []string{"$1"}
	for i := range t.Columns {
		cols = append(cols, fmt.Sprintf("$%d", i+2))
	}
	return strings.Join(cols, ", ")
}

func (t TableMap) PKCol() string {
	return "id"
}

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
}

type Code struct {
	buf  *bytes.Buffer
	tmpl *template.Template
	dest string
}

func (c Code) write(format string, param ...interface{}) {
	c.buf.WriteString(fmt.Sprintf(format, param...))
}

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
		if _, err := c.buf.WriteTo(f); err != nil {
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
		tmpl: template.Must(template.ParseFiles("tablestruct.go.tmpl")),
		dest: flag.Arg(1),
	}
	code.Gen(mapper, "main")
}
