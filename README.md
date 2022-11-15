# db-schema

db-schema is a command line utility for extracting schema from SQLite databases.

## Installation

```bash
go install github.com/adnsv/db-schema@latest
```

## Usage

Retrieving chema from an existing SQLite database file:


```txt
db-shema scan [flags] DBFILE

The flags are:

  -fmt <json|yaml|sql>
        output data format (json|yaml|sql)
  -normnames <upper|lower>
        normalize names tables/columns/indices to upper or lower case
  -normtypes
        normalize column types, remove default nulls for nullables
  -output <filepath>
        output filename
  -sort <tables/columns/indices>
        sort output tables/columns/indices
```

Scans SQLite database file for existing schema.

Specify output filename with the `-output=<filename>` flag, 
otherwise the output is dumped to stdout.

The extracted schema can be dumped in json, yaml, or sql formats 
(guessed from the output file extension, defaults to json 
for console output).

Override format with `-fmt=json` `-fmt=yaml` or `-fmt=sql` flags.

With the `-normtypes` flag, column types can be normalized:

original | normalized
---------|-----------
int integer tinyint smallint mediumint   | int
int64 bigint                             | int64
boolean bool                             | bool
real double float                        | float
blob                                     | blob
text string clob                         | text
date                                     | date
time                                     | time
datetime timestamp                       | timestamp
uuid, guid                               | uuid
character() varchar() nchar() nvarchar() | text

Notice, that `-normtypes` also removed `default null` from nullable
columns (that don't have `not null` in their type).

With the `-normnames=<upper|lower>` flag, the names for all
the tables, columns, and indices can be converted to upper or lower
case.

With the `-sort` flag, tables/columns/indices can be sorted:

flag | action
-----|-------
-sort=tables                 | sorts tables by name
-sort=columns,indices        | sorts content within each table
-sort=tables,columns,indices | sorts everything (nice for diffing)

Note: sorting is performed lexicographically. If `-normnames` is 
also specified, sorting is performed after all the name normalization.