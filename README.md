tablestruct
===========

[![Build Status](https://travis-ci.org/paulsmith/tablestruct.svg)](https://travis-ci.org/paulsmith/tablestruct)

`tablestruct` maps Go structs to database tables, and struct fields to columns.

It is a lightweight alternative to ORMs. It preserves type-safety and eschews reflection. 

It provides common operations for persisting structs to and retrieving
structs from the database, including but not limited to:

* get by ID/primary key
* insert
* insert many (in a transaction)
* update
* delete
* find by WHERE clause

Current release: **0.1.0** (July 19, 2014)

Tentative roadmap
-----------------

* 0.2: Inspect db to generate initial mapping metadata file.
* 0.3: Generate structs from metadata.
* 0.4: Generate db schema from metadata.
* 0.5: Foreign keys/relationships between structs.

tablestruct documentation
=========================

Installation
------------

Requires Go 1.2.

```bash
go get github.com/paulsmith/tablestruct/cmd/tablestruct
```

Quickstart
----------

In the following, let's assume you want to map a Go struct type named `Person`,
in a file named `person.go` already on disk, to a database table named `people`.

```bash
$ go get github.com/paulsmith/tablestruct/cmd/tablestruct
$ tablestruct support > mapper_support.go # one-time support code generation step
$ tablestruct struct Person < person.go > person_mapper.metadata
$ $EDITOR person_mapper.metadata # tweak to adjust for differences between struct and table, column and field names
$ tablestruct gen < person_mapper.metadata > person_mapper.go
```

You should also `go get` your database's driver if you haven't already.

```bash
$ go get github.com/lib/pq # for PostgreSQL, only supported db in tablestruct so far ...
```

Motivation
----------

If you work with Go and databases regularly, and you have Go structs which
correspond to or are persisted by database tables, you might find yourself
writing a lot of the same code over and over, tedious bits that map struct
fields to columns, templatize SQL, perform the equivalent of inserts and updates
but on structs instead of rows, and so forth.

You can write much of this code once and rely on reflection, but it has its
costs, in loss of compile-time type safety and increased latency at runtime.
Reflection-based mappers tend to have complicated implementations as well.

tablestruct is meant to ease the workaday usage and maintenance of structs that
map to tables. It is a goal to have a simple implementation, the intermediate
steps of which are easily inspectable.

### Ambition

tablestruct has low ambition. It does not intend to do that much, for example,
it will never be a complete ORM, nor will it ever replace writing SQL
inside Go code. It merely wants to be useful for a set of common, low-level data
persistence operations.

Theory of operation
-------------------

tablestruct has three separate stages of operation, but the first one only needs
to be run once initially, and the second is optional or otherwise only needs to
be run once. The third stage is the code generation step that is the meat of
tablestruct.

tablestruct produces mappers, which are structs that handle the back and forth
between the main structs in your Go application and the database. Specifically,
they handle the mapping between structs and tables, and fields and columns, in
SQL for operations like inserts and updates, and in Go code for initializing new
structs from query result sets. The mappers are generated as separate files
of Go code that reside in the same package as the rest of your code. You can
be override this behavior to target any package.

The way to use tablestruct during the normal course of development is to create
your main structs and your database tables, run tablestruct as a command to
produce the mapper files, and then `go run` or `go build`.

You should run tablestruct's code gen stage after any time you modify your main
structs or the database tables. This is best handled by making the code
generation step a rule in a Makefile, discussed below. Generated Go code should
not be checked in to version control.

Current limitations
-------------------

tablestruct currently only supports PostgreSQL. This is merely due to it being
the database it was initially developed against, and is not a limitation of its
design. To support other databases, a mild refactoring would be needed that
allowed for the code generation template to be parameterized.

Mapping metadata
----------------

The information that enables tablestruct to generate the correct mapper code is
the mapping metadata. The metadata can be a file that you edit by hand, or it
can be piped in from one stage of tablestruct to another with no intermediate
files produced, if you are happy with the translation between field names and
column names, as well as struct type names and table names.

Mapping metadata is encoded as JSON. You can pass it in as stdin to the code gen
stage of tablestruct.

If you are just starting out, you can generate initial metadata from your
existing Go structs. Run `tablestruct metadata`, passing the name of the struct
type you want to map as an argument, and pipe the `.go` file containing the
struct as stdin:

```bash
$ tablestruct metadata Person < person.go > person.metadata
```

tablestruct will try to convert exported Go struct type and field names into
database table and column names (broadly, going from `CamelCase` to
`snake_case`). However, you may have less simplistic translations, for example
Go struct type `Person` and table name `people`, so tablestruct doesn't attempt
to try to outsmart you. It is intended that you will edit the metadata file by
hand to get the exact name translations right.

Mapper API
----------

Assume `T` is the name of your main Go struct type that is mapped.

```go
func NewTMapper(db *sql.DB) *TMapper
```

```go
func (m *TMapper) Insert(t *T) error
```

```go
func (m *TMapper) InsertMany(t []*T) error
```

```go
func (m *TMapper) Update(t *T) error
```

```go
func (m *TMapper) Get(id int64) (*T, error)
```

```go
func (m *TMapper) All() ([]*T, error)
```

```go
func (m *TMapper) FindWhere(whereClause string) ([]*T, error)
```

```go
func (m *TMapper) Delete(t *T) error
```

Example
-------

Say you've got a struct type that's your model in your application:

```go
// event.go

package mypkg

import "database/sql"

type Event struct {
    ID int64
    Title string
    Description sql.NullString
    UnixTimestamp int64
    Severity float64
    Resolved bool
}
```

And you have a corresponding table in a database where structs are persisted as
rows:

```sql
create table event (
    id bigserial primary key,
    title varchar not null,
    description varchar,
    unix_timestamp bigint not null default now,
    severity double precision not null default 0.0,
    check (severity >= 0.0 and severity <= 1.0),
    resolved bool not null default 'f'
);
```

Generate the mapper:

```bash
$ tablestruct -package mypkg support > mapper_support.go
$ tablestruct -package mypkg metadata Event < event.go | tablestruct -pkg mypkg gen > event_mapper.go
```

Look at the mapper:

```go
// event_mapper.go

// generated mechanically by tablestruct, do not edit!!
package mypkg

import (
	"database/sql"
	"log"
)

type EventMapper struct {
	db   *sql.DB
	sql  map[string]string
	stmt map[string]*sql.Stmt
}

func NewEventMapper(db *sql.DB) *EventMapper {
	m := &EventMapper{
		db:   db,
		sql:  make(map[string]string),
		stmt: make(map[string]*sql.Stmt),
	}
	m.prepareStatements()
	return m
}

// ... lots of code elided ...
```

Write some application code to use it:

```go
// main.go

package main

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/lib/pq"
    "path/to/mypkg"
)

func main() {
    db, _ := sql.Open("postgres", "")

    mapper := mypkg.NewEventMapper(db)

    event := &mypkg.Event{
        Title: "my severe event",
        UnixTimestamp: time.Now().Unix(),
        Severity: 1.0,
    }

    mapper.Insert(event)

    event.Resolved = true
    event.Description = sql.NullString{String: "investigated and resolved", Valid: true}
    mapper.Update(event)

    fmt.Println(event.ID)
    // Output: 1

    events := []*mypkg.Event{
        {
            Title: "my not severe event",
            UnixTimestamp: time.Now().Unix(),
            Severity: 0.0,
        },
        {
            Title: "my slightly severe event",
            UnixTimestamp: time.Now().Unix(),
            Severity: 0.3,
        },
        {
            Title: "my moderately severe event",
            UnixTimestamp: time.Now().Unix(),
            Severity: 0.5,
        },
    }

    mapper.InsertMany(events)

    events, _ = mapper.FindWhere("severity < 0.5")
    fmt.Println(len(events))
    // Output: 2
}
```

Tips & tricks
-------------

### Make

The best way to use tablestruct is via `make(1)` and `Makefiles`.

Set up a rule where the target is the generated mapper file and the prerequisite
is the Go source file containing the struct type:

```make
PACKAGE=main

event_mapper: event.go
    tablestruct -package=$(PACKAGE) metadata Event < $< | \
    tablestruct -package=$(PACKAGE) gen > $@
```

You'll need the mapper support file, so good to add that:

```make
.PHONY: mapper_support.go
mapper_support.go:
    tablestruct -package=$(PACKAGE) support > $@
```

If you have multiple struct files, it is better to make use of `make`'s pattern
rules. The only trick is to match the struct type name to the Go file it's in.

```make
STRUCT_event=Event
STRUCT_person=Person

%_mapper.go: %.go
    tablestruct -package=$(PACKAGE) metadata $(STRUCT_$(basename $<)) < $< | \
    tablestruct -package=$(PACKAGE) gen > $@
```

Then gather up all the mapper files as a separate target:

```make
struct_files := event.go person.go
mapper_files := $(patsubst %.go,%_mapper.go,$(struct_files))

all: $(mapper_files)
```

The final `Makefile` looks something like:

```make
PACKAGE=main

STRUCT_event=Event
STRUCT_person=Person

struct_files := event.go person.go
mapper_files := $(patsubst %.go,%_mapper.go,$(struct_files))

all: $(mapper_files) mapper_support.go

%_mapper.go: %.go
    tablestruct -package=$(PACKAGE) metadata $(STRUCT_$(basename $<)) < $< | \
    tablestruct -package=$(PACKAGE) gen > $@

.PHONY: mapper_support.go
mapper_support.go:
    tablestruct -package=$(PACKAGE) support > $@
```

Then run `make`:

```bash
$ make PACKAGE=mypkg
tablestruct -package=mypkg metadata Event < event.go | \
	tablestruct -package=mypkg gen > event_mapper.go
tablestruct -package=mypkg metadata Person < person.go | \
	tablestruct -package=mypkg gen > person_mapper.go
tablestruct -package=mypkg support > mapper_support.go
```
