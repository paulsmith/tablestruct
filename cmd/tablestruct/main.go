package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/paulsmith/tablestruct"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-package=<package>] gen\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "   or: %s [-package=<package>] [-table=<table>] [-pk=<field>] metadata <structname>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "   or: %s [-package=<package>] support\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "option defaults:\n")
	flag.PrintDefaults()
}

type command struct {
	name string // how it is invoked on the command line
	fn   func()
}

type commands []command

func (c commands) invoke(name string) {
	for i := range c {
		if c[i].name == name {
			c[i].fn()
			break
		}
	}
}

// Generate Go code from mapping metadata.
func gen(pkg string) {
	mapper, err := tablestruct.NewMap(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	code := tablestruct.NewCode()
	code.Gen(mapper, pkg, os.Stdout)
}

// Generate metadata by inspecting a struct.
func structMetadata(typ, overrideTable, pkField string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", os.Stdin, 0)
	if err != nil {
		log.Fatal(err)
	}

	var mapper tablestruct.Map

	// Find the struct type named `typ' in the AST.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name != typ {
				break
			}

			structType, ok := x.Type.(*ast.StructType)
			if !ok {
				log.Fatalf("%q must be a struct type, got %T", typ, x.Type)
			}

			tableName := overrideTable
			if tableName == "" {
				tableName = tablestruct.StructToTable(typ)
			}

			tableMap := tablestruct.TableMap{
				Struct:  typ,
				Table:   tableName,
				Columns: make([]tablestruct.ColumnMap, 0, len(structType.Fields.List)),
				// TODO(paulsmith): allow override
				AutoPK: false,
			}

			for i, field := range structType.Fields.List {
				if field.Names == nil {
					continue
				}
				name := field.Names[0]
				if !name.IsExported() {
					continue
				}
				ident, ok := field.Type.(*ast.Ident)
				if !ok {
					log.Printf("field %d %q is anonymous type dec, skipping", i, ident)
					continue
				}
				column := tablestruct.ColumnMap{
					Field:  name.Name,
					Column: tablestruct.FieldToColumn(name.Name),
					//Type: tablestruct.StructTypeToColumnType(fieldType),
					Null:       false,
					PrimaryKey: name.Name == pkField,
				}
				tableMap.Columns = append(tableMap.Columns, column)
			}

			mapper = tablestruct.Map{tableMap}
		}

		return true
	})

	if err := json.NewEncoder(os.Stdout).Encode(mapper); err != nil {
		log.Fatal(err)
	}
}

// Generate supporting Go code.
func support(pkg string) {
	tablestruct.GenSupport(os.Stdout, pkg)
}

func main() {
	var (
		pkg           = flag.String("package", "main", "package of generated code")
		overrideTable = flag.String("table", "", "override table name")
		pkField       = flag.String("pk", "ID", "name of struct field of primary key")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "must supply subcommand\n")
		flag.Usage()
		os.Exit(1)
	}

	cmds := commands{
		{"gen", func() { gen(*pkg) }},
		{"metadata", func() {
			if flag.Arg(1) == "" {
				fmt.Fprintf(os.Stderr, "must supply name of struct type\n")
				flag.Usage()
				os.Exit(1)
			}
			structMetadata(flag.Arg(1), *overrideTable, *pkField)
		},
		},
		{"support", func() { support(*pkg) }},
	}

	cmds.invoke(flag.Arg(0))
}
