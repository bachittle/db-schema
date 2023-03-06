package main

import (
	dbs "github.com/adnsv/db-schema"
	"github.com/alecthomas/kong"
)

var cli struct {
	Scan     dbs.ScanCmd          `cmd:"" help:"Retrieve schema from the database file."`
	Compare  dbs.CompareCmd       `cmd:"" help:"Compare database schemas, generate a migration file."`
	Version  kong.VersionFlag     `short:"v" help:"Print version information and quit."`
	Migrator dbs.MigrationToolCmd `cmd:"" help:"Migration tool for sqlite databases."`
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
