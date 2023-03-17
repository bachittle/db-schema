// database schema migrator
// migrate the schema if it is trivial (no data loss)
// otherwise, perform the migrations manually
package main

import (
	"fmt"
)

type MigrateCmd struct {
	Src string `arg:"" type:"existingfile" help:"Source database file that will be migrated. (sqlite3)"`
	Dst string `arg:"" type:"existingfile" help:"Destination database file that the src will be converted to. (sqlite3, yaml)"`
}

func (v *MigrateCmd) Run() error {
	fmt.Println("# migrate", v.Src, v.Dst)

	// load src and dst
	// srcData, err := os.ReadFile(v.Src)
	// if err != nil {
	// 	return err
	// }

	// dstData, err := os.ReadFile(v.Dst)
	// if err != nil {
	// 	return err
	// }

	return nil
}
