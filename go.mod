module github.com/adnsv/db-schema

go 1.19

require (
	github.com/adnsv/go-db3 v0.5.0
	github.com/alecthomas/kong v0.7.1
	github.com/mattn/go-sqlite3 v1.14.16
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/exp v0.0.0-20221111094246-ab4555d3164f // indirect

replace github.com/adnsv/go-db3 => ../go-db3
