tablestruct
===========

[![Build Status](https://travis-ci.org/paulsmith/tablestruct.svg)](https://travis-ci.org/paulsmith/tablestruct)

Maps Go structs to database tables, struct fields to columns.

It provides common functionality for persisting structs to and retrieving
structs from the database, such as:

* get by ID/primary key
* insert
* insert many (in transaction)
* update
* delete
* find by WHERE clause

tablestruct uses code generation.

Maintain a metadata file that contains the mappings. This file can be initially
created by inspecting an existing database or set of tables.

Generated Go code should not be checked in to version control. The generation
step should be included as part of a build process.

Roadmap
-------

* 0.1: Initial release - basic mapping functionality.
* 0.2: Generate mapping metadata from existing structs.
* 0.3: Inspect db to generate initial mapping metadata file.
* 0.4: Generate structs from metadata.
* 0.5: Generate db schema from metadata.
* 0.6: Joins.
