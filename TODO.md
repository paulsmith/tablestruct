* [ ] Include SQL functions or other expressions in output rows
* [ ] Handle aggregates (don't?)
* [ ] Handle LIMIT and OFFSET
* [ ] Handle ORDER BY
* [ ] Figure out solution for FindWhere that's SQL-injection-safe
* [ ] update multiple
* [ ] delete multiple
* [ ] exists
* [ ] count (don't?)
* [ ] find by field
* [ ] find one by field
* [ ] save (insert/update)
* [ ] override naming
* [ ] Hooks for adding custom code
* [ ] Factory to get a mapper for a struct (registry?)
* [ ] Support transactions
* [ ] Other dialects (MySQL, SQLite) - main thing is "RETURNING" syntax on INSERT
  [ ] stmts
* [ ] Un-export things
* [ ] Helpers for when you just want to write SQL
* [ ] Rename VarName -> Var or something shorter in template
* [ ] Clean up examples
* [ ] Write up getting started
* [ ] Async APIs
* [ ] Low-level Query API
* [ ] Factor out "type"writers
* [ ] Move table name to mapper struct
* [ ] Embed *sql.DB in mapper structs
* [ ] Exclude methods (ones you don't need like Delete(), etc.)
* [ ] Move pk bool from column to table
* [ ] Allow time.Time, etc. to be type for field (currently thinks is anon struct)
* [ ] If `metadata` and no struct name is given, and if only one exported struct type
  [ ] in file, use that
* [ ] Remove importing lib/pq from codegen
* [ ] Clean up genMapper
* [ ] Move "Mapper" prefix to Code field and let main driver set it
