package main

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
        "Find": "SELECT {{.ColumnList}} FROM {{.Table}} WHERE {{.PKCol}} = $1",
        "Update": "UPDATE {{.Table}} SET {{.UpdateList}} WHERE {{.PKCol}} = $1",
        "Insert": "INSERT INTO {{.Table}} VALUES ({{.InsertList}})",
        "Delete": "DELETE FROM {{.Table}} WHERE {{.PKCol}} = $1",
    }
    for k, v := range rawSql {
        stmt, err := {{.VarName}}.db.Prepare(v)
        if err != nil {
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

func ({{.VarName}} {{.MapperType}}) Find(key int64) (*{{.StructType}}, error) {
    row := {{.VarName}}.stmt["Find"].QueryRow(key)
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

func ({{.VarName}} {{.MapperType}}) Insert(obj *{{.StructType}}) error {
    args := []interface{}{
        {{range .Fields}}obj.{{.}},
        {{end}}
    }
    _, err := {{.VarName}}.stmt["Insert"].Exec(args...)
    return err
}

func ({{.VarName}} {{.MapperType}}) InsertMany(objs []*{{.StructType}}) error {
    tx, err := {{.VarName}}.db.Begin()
    if err != nil {
        return err
    }
    stmt := tx.Stmt({{.VarName}}.stmt["Insert"])
    for _, obj := range objs {
        args := []interface{}{
            {{range .Fields}}obj.{{.}},
            {{end}}
        }
        _, err := stmt.Exec(args...)
        if err != nil {
            return err
        }
    }
    return tx.Commit()
}

func ({{.VarName}} {{.MapperType}}) FindWhere(where string) ([]*{{.StructType}}, error) {
    sql := "SELECT {{.ColumnList}} FROM {{.Table}} WHERE " + where
    rows, err := {{.VarName}}.db.Query(sql)
    if err != nil {
        return nil, err
    }
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

func ({{.VarName}} {{.MapperType}}) Delete(obj *{{.StructType}}) error {
    _, err := {{.VarName}}.stmt["Delete"].Exec(obj.{{.PKField}})
    return err
}

{{end}}
`
