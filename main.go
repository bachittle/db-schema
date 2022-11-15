package main

import (
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	cc := &cli_context{
		descr:      "schema scanner for sqlite databases",
		executable: "db-shema",
		commands: map[string]cmd{
			"version": &version_cmd{},
			"scan":    &scan_cmd{},
		},
	}

	err := cc.execute(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
