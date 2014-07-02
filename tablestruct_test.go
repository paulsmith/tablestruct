package main

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

type Fataler interface {
	Fatal(...interface{})
}

func openTestDB(f Fataler) *sql.DB {
	dbname := os.Getenv("PGDATABASE")
	sslmode := os.Getenv("PGSSLMODE")
	timeout := os.Getenv("PGCONNECT_TIMEOUT")

	if dbname == "" {
		os.Setenv("PGDATABASE", "tablestruct_test")
	}

	if sslmode == "" {
		os.Setenv("PGSSLMODE", "disable")
	}

	if timeout == "" {
		os.Setenv("PGCONNECT_TIMEOUT", "10")
	}

	db, err := sql.Open("postgres", "")
	if err != nil {
		f.Fatal(err)
	}
	return db
}

func tempDir(f Fataler) string {
	name, err := ioutil.TempDir("", "tablestruct_test")
	if err != nil {
		f.Fatal(err)
	}
	return name
}

func tempFile(dir string, f Fataler) *os.File {
	file, err := ioutil.TempFile(dir, "tablestruct_test")
	if err != nil {
		f.Fatal(err)
	}
	return file
}

func tempGoFile(dir string, f Fataler) *os.File {
	file := tempFile(dir, f)
	os.Symlink(file.Name(), file.Name()+".go")
	return file
}

var get = CodeGenTest{
	CreateTableSQL: `CREATE TABLE t (id int, val int)`,
	CleanupSQL:     `DROP TABLE t`,
	TableSetupSQL:  `INSERT INTO t (SELECT generate_series(0, 10), generate_series(100, 110))`,
	Metadata: `
[
    {
        "struct": "T",
        "table": "t",
        "columns": [{
            "field": "Value",
            "column": "val",
            "type": "int"
        }]
    }
]
`,
	DriverCode: `
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

type T struct {
    ID    int
    Value int
}

func main() {
    db, err := sql.Open("postgres", "")
    if err != nil {
        log.Fatal(err)
    }
    m := NewTMapper(db)
    t, err := m.Get(8)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d\n", t.Value)
}
`,
	Expected: "108\n",
}

var all = CodeGenTest{
	CreateTableSQL: `CREATE TABLE zipcodes (id int, zipcode varchar)`,
	CleanupSQL:     `DROP TABLE zipcodes`,
	TableSetupSQL:  `INSERT INTO zipcodes (SELECT generate_series(0, 9), generate_series(21230, 21239)::varchar)`,
	Metadata: `
[
    {
        "struct": "ZIPCode",
        "table": "zipcodes",
        "columns": [{
            "field": "Z5",
            "column": "zipcode",
            "type": "varchar"
        }]
    }
]
`,
	DriverCode: `
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

type ZIPCode struct {
    ID    int
    Z5    string
}

func main() {
    db, err := sql.Open("postgres", "")
    if err != nil {
        log.Fatal(err)
    }
    m := NewZIPCodeMapper(db)
    zips, err := m.All()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d\n", len(zips))
    for i := range zips {
        fmt.Printf("%s\n", zips[i].Z5)
    }
}
`,
	Expected: `10
21230
21231
21232
21233
21234
21235
21236
21237
21238
21239
`}

var insert = CodeGenTest{
	CreateTableSQL: `CREATE TABLE person (id int, name varchar, age int)`,
	CleanupSQL:     `DROP TABLE person`,
	Metadata: `
[
    {
        "struct": "Person",
        "table": "person",
        "columns": [{
            "field": "Name",
            "column": "name",
            "type": "varchar"
        }, {
            "field": "Age",
            "column": "age",
            "type": "int"
        }]
    }
]
`,
	DriverCode: `
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/lib/pq"
)

type Person struct {
    ID      int64
    Name    string
    Age     int
}

func main() {
    db, err := sql.Open("postgres", "")
    if err != nil {
        log.Fatal(err)
    }
    m := NewPersonMapper(db)
    p := Person{42, "Paul Smith", 37}
    err = m.Insert(&p)
    if err != nil {
        log.Fatal(err)
    }
    dest := []interface{}{
        new(int64),
        new(string),
        new(int),
    }
    err = db.QueryRow("SELECT * FROM person WHERE id = 42").Scan(dest...)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d '%s' %d\n", *dest[0].(*int64), *dest[1].(*string), *dest[2].(*int))
}
`,
	Expected: "42 'Paul Smith' 37\n",
}

type CodeGenTest struct {
	CreateTableSQL string
	TableSetupSQL  string
	CleanupSQL     string
	Metadata       string
	DriverCode     string
	Expected       string
}

func testCodeGen(t *testing.T, test CodeGenTest) {
	db := openTestDB(t)
	defer db.Close()

	_, err := db.Exec(test.CreateTableSQL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if test.CleanupSQL != "" {
			db.Exec(test.CleanupSQL)
		}
	}()

	if test.TableSetupSQL != "" {
		_, err = db.Exec(test.TableSetupSQL)
		if err != nil {
			t.Fatal(err)
		}
	}

	mapper, err := NewMap(strings.NewReader(test.Metadata))
	if err != nil {
		t.Fatal(err)
	}

	dir := tempDir(t)
	genCodeFile := tempGoFile(dir, t)
	driverCodeFile := tempGoFile(dir, t)

	defer func() {
		genCodeFile.Close()
		driverCodeFile.Close()
		os.RemoveAll(dir)
	}()

	code := NewCode()
	code.Gen(mapper, "main", genCodeFile)

	if _, err := driverCodeFile.WriteString(test.DriverCode); err != nil {
		t.Fatal(err)
	}

	genCodeFile.Sync()
	driverCodeFile.Sync()

	goFiles, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	args := []string{"run"}
	for i := range goFiles {
		args = append(args, goFiles[i])
	}

	cmd := exec.Command("go", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Fatalf("error running generated Go code: %s", err)
	}

	if actual := stdout.String(); actual != test.Expected {
		t.Errorf("want %q, got %q", test.Expected, actual)
	}
}

func TestCodeGen(t *testing.T) {
	var tests = map[string]CodeGenTest{
		"Get":    get,
		"All":    all,
		"Insert": insert,
	}
	for name, test := range tests {
		t.Log(name)
		testCodeGen(t, test)
	}
}
