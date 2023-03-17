package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	Scan    ScanCmd          `cmd:"" help:"Retrieve schema from the database file."`
	Compare CompareCmd       `cmd:"" help:"Compare database schemas, generate a migration file."`
	Migrate MigrateCmd       `cmd:"" help:"Migrate the database schema."`
	Version kong.VersionFlag `short:"v" help:"Print version information and quit."`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("db-schema"),
		kong.Description("Schema scanner for sqlite databases."),
		kong.UsageOnError(),
		kong.Vars{"version": app_version()},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
