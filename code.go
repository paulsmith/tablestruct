package tablestruct

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"strings"
	"text/template"
)

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
