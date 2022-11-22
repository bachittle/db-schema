# db-schema

db-schema is a command line utility for extracting schema from SQLite databases.

## Installation

```bash
go install github.com/adnsv/db-schema@latest
```

## Usage

Retrieving chema from an existing SQLite database file:


```txt
Usage: db-schema scan [flags] <input-file>

Arguments:
  <input-file>    Path to SQLite database file.

Flags:
  -h, --help                      Show context-sensitive help.
  -v, --version                   Print version information and quit.

  -o, --output=STRING             Output filename.
  -f, --fmt=yaml|json|sql         Output format.
      --norm-types                Normalize column types, remove default nulls for nullables.
      --norm-names=upper|lower    Normalize names to upper/lower case.
  -s, --sort=PART,...             Sort output (tables/columns/indices).
```

Scans SQLite database file for existing schema.

Specify output filename with the `--output=FILENAME` flag, 
otherwise the output is dumped to stdout.

The extracted schema can be dumped in json, yaml, or sql formats 
(guessed from the output file extension, defaults to json 
for console output).

Override format with `--fmt=json` `--fmt=yaml` or `--fmt=sql` flags.

With the `--norm-types` flag, column types can be normalized:

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

Notice, that `--norm-types` also removes `default null` from nullable
columns (that don't have `not null` in their type).

With the `--norm-names=upper|lower` flag, the names for all
the tables, columns, and indices can be converted to upper or lower
case.

With the `--sort` flag, tables, columns, and/or indices can be sorted:

flag | action
-----|-------
`--sort=tables`                 | sorts tables by name
`--sort=columns,indices`        | sorts content within each table
`--sort=tables,columns,indices` | sorts everything (nice for diffing)

Note: sorting is performed lexicographically. If `--norm-names` is 
also specified, sorting is performed after all the name normalization.

## License

The db-schema utility is licenced under the MIT license 

Other libraries used:
- https://github.com/alecthomas/kong
- https://github.com/adnsv/go-db3
- https://github.com/mattn/go-sqlite3
- https://github.com/go-yaml/yaml
- https://github.com/wangyoucao577/go-release-action