package main

var mapperTemplate = `
// generated mechanically by tablestruct, do not edit!!
package {{.Package}}

import (
    {{range .Imports}}{{.}}
    {{end}}
)

type {{.MapperType}} struct {
    {{range .MapperFields}}{{.}}
    {{end}}
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
    sql := "SELECT {{.ColumnList}} FROM {{.Table}} WHERE {{.PKCol}} = $1"
    row := {{.VarName}}.db.QueryRow(sql, key)
    return {{.VarName}}.loadObj(row)
}

func ({{.VarName}} {{.MapperType}}) Update(obj *{{.StructType}}) error {
    sql := "UPDATE {{.Table}} SET {{.UpdateList}} WHERE {{.PKCol}} = $1"
    args := []interface{}{
        {{range .Fields}}obj.{{.}},
        {{end}}
    }
    _, err := {{.VarName}}.db.Exec(sql, args...)
    return err
}

func ({{.VarName}} {{.MapperType}}) Insert(obj *{{.StructType}}) error {
    sql := "INSERT INTO {{.Table}} VALUES ({{.InsertList}})"
    args := []interface{}{
        {{range .Fields}}obj.{{.}},
        {{end}}
    }
    _, err := {{.VarName}}.db.Exec(sql, args...)
    return err
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
`
