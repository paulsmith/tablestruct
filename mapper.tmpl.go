package tablestruct

var mapperTemplate = `
// generated mechanically by tablestruct, do not edit!!
package {{.Package}}

import (
    {{range .Imports}}{{.}}
    {{end}}
)

type Scanner interface {
    Scan(...interface{}) error
}

{{range .TableMaps}}

type {{.MapperType}} struct {
    {{range .MapperFields}}{{.}}
    {{end}}
}

func New{{.MapperType}} (db *sql.DB) *{{.MapperType}} {
    m := &{{.MapperType}}{
        db: db,
        sql: make(map[string]string),
        stmt: make(map[string]*sql.Stmt),
    }
    m.prepareStatements()
    return m
}

func ({{.VarName}} {{.MapperType}}) prepareStatements() {
    var rawSql = map[string]string{
        "Get": "SELECT {{.ColumnList}} FROM {{.Table}} WHERE {{.PKCol}} = $1",
        "Update": "UPDATE {{.Table}} SET {{.UpdateList}} WHERE {{.PKCol}} = $1",
        "Insert": "INSERT INTO {{.Table}} VALUES ({{.Mapper.InsertList}}) RETURNING {{.PKCol}}",
        "Delete": "DELETE FROM {{.Table}} WHERE {{.PKCol}} = $1",
        "All": "SELECT {{.ColumnList}} FROM {{.Table}}",
    }
    for k, v := range rawSql {
        stmt, err := {{.VarName}}.db.Prepare(v)
        if err != nil {
            // TODO(paulsmith): return error instead.
            log.Printf("SQL: %q", v)
            log.Fatalf("preparing %s SQL: %v", k, err)
        }
        {{.VarName}}.stmt[k] = stmt
        {{.VarName}}.sql[k] = v
    }
}

func ({{.VarName}} {{.MapperType}}) loadObj(scanner Scanner) (obj *{{.StructType}}, err error) {
    obj = new({{.StructType}})
    dest := []interface{}{
        {{range .Fields}}&obj.{{.}},
        {{end}}
    }
    err = scanner.Scan(dest...)
    return
}

func ({{.VarName}} {{.MapperType}}) Get(key int64) (*{{.StructType}}, error) {
    row := {{.VarName}}.stmt["Get"].QueryRow(key)
    return {{.VarName}}.loadObj(row)
}

func ({{.VarName}} {{.MapperType}}) Update(obj *{{.StructType}}) error {
    args := []interface{}{
        {{range .Fields}}obj.{{.}},
        {{end}}
    }
    _, err := {{.VarName}}.stmt["Update"].Exec(args...)
    return err
}

func ({{.VarName}} {{.MapperType}}) insert(obj *{{.StructType}}, stmt *sql.Stmt) error {
    args := []interface{}{
        {{range .Mapper.InsertFields}}obj.{{.}},
        {{end}}
    }
    row := stmt.QueryRow(args...)
    err := row.Scan(&obj.{{.PKField}})
    return err
}

func ({{.VarName}} {{.MapperType}}) Insert(obj *{{.StructType}}) error {
    return {{.VarName}}.insert(obj, {{.VarName}}.stmt["Insert"])
}

func ({{.VarName}} {{.MapperType}}) InsertMany(objs []*{{.StructType}}) error {
    tx, err := {{.VarName}}.db.Begin()
    if err != nil {
        return err
    }
    stmt := tx.Stmt({{.VarName}}.stmt["Insert"])
    for _, obj := range objs {
        err := {{.VarName}}.insert(obj, stmt)
        if err != nil {
            return err
        }
    }
    return tx.Commit()
}

func ({{.VarName}} {{.MapperType}}) loadManyObjs(rows *sql.Rows) ([]*{{.StructType}}, error) {
    var objs []*{{.StructType}}
    for rows.Next() {
        obj, err := {{.VarName}}.loadObj(rows)
        if err != nil {
            return nil, err
        }
        objs = append(objs, obj)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return objs, nil
}

func ({{.VarName}} {{.MapperType}}) FindWhere(where string) ([]*{{.StructType}}, error) {
    sql := "SELECT {{.ColumnList}} FROM {{.Table}} WHERE " + where
    rows, err := {{.VarName}}.db.Query(sql)
    if err != nil {
        return nil, err
    }
    return {{.VarName}}.loadManyObjs(rows)
}

func ({{.VarName}} {{.MapperType}}) All() ([]*{{.StructType}}, error) {
    rows, err := {{.VarName}}.stmt["All"].Query()
    if err != nil {
        return nil, err
    }
    return {{.VarName}}.loadManyObjs(rows)
}

func ({{.VarName}} {{.MapperType}}) Delete(obj *{{.StructType}}) error {
    _, err := {{.VarName}}.stmt["Delete"].Exec(obj.{{.PKField}})
    return err
}

func ({{.VarName}} {{.MapperType}}) Table() string {
    return "{{.Table}}"
}

{{end}}
`
