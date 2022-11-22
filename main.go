package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	Scan    scan_cmd         `cmd:"" help:"Retrieve schema from the database file."`
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
