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
	file, err := ioutil.TempFile(dir, "tablestruct_test_Find")
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

func TestFind(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, err := db.Exec(`CREATE TABLE t (id int, val int)`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		// TODO(paulsmith): create temp table doesn't work here, need to
		// investigate more.
		db.Exec(`DROP TABLE t`)
	}()

	_, err = db.Exec(`INSERT INTO t (SELECT generate_series(0, 10), generate_series(100, 110))`)
	if err != nil {
		t.Fatal(err)
	}

	var metadata = `
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
`
	mapper, err := NewMap(strings.NewReader(metadata))
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

	var driverCode = `
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
    t, err := m.Find(8)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d\n", t.Value)
}
`
	if _, err := driverCodeFile.WriteString(driverCode); err != nil {
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
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	expected := "108\n"
	if actual := out.String(); actual != expected {
		t.Errorf("want %q, got %q", expected, actual)
	}
}
