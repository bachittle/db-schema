// database schema migrator
// migrate the schema if it is trivial (no data loss)
// otherwise, perform the migrations manually
package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/adnsv/go-db3/schema"
)

type MigrateCmd struct {
	Src string `arg:"" type:"existingfile" help:"Source database file that will be migrated. (sqlite3)"`
	Dst string `arg:"" type:"existingfile" help:"Destination database file as a reference that will convert source. (sqlite3, yaml)"`
}

func (v *MigrateCmd) Run() error {
	fmt.Println("# migrate", v.Src, v.Dst)

	db1, err := sql.Open("sqlite3", v.Src)
	if err != nil {
		return err
	}
	defer db1.Close()

	db2, err := sql.Open("sqlite3", v.Dst)
	if err != nil {
		return err
	}
	defer db2.Close()

	// convert db1 => db2
	// writes on db1, db2 is the reference

	sch1, err := schema.Scan(db1)
	if err != nil {
		return err
	}

	sch2, err := schema.Scan(db2)
	if err != nil {
		return err
	}

	migration := compareSchemas(sch1.Tables, sch2.Tables)

	// perform migration

	// trivial if empty
	if len(migration.Add) == 0 && len(migration.Rm) == 0 {
		fmt.Println("Database is up to date.")
		return nil
	}

	// remove tables first. this can cause data loss
	for _, table := range migration.Rm {
		// ask for confirmation
		var confirm string
		fmt.Printf("Are you sure you want to remove table %s? (y/n): ", table.Name)
		fmt.Scanln(&confirm)

		if confirm == "y" {
			_, err := db1.Exec(fmt.Sprintf("DROP TABLE %s;", table.Name))
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Done removing %d tables.\n", len(migration.Rm))

	// add new tables
	for _, table := range migration.Add {
		// we don't need confirmation for this, since there is no data loss

		// check if table already exists
		var count int
		err := db1.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table.Name).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			// create the table

			// find id column
			idCol, found := table.FindColumn("id")
			if !found {
				panic(fmt.Sprintf("%s doesn't have an id column", table.Name))
			}

			// create table

			// TODO: we cannot use ? here because the column name is not a parameter
			// does this mean that we are vulnerable to SQL injection?
			// how do we prevent this?
			log.Printf("sql: CREATE TABLE %s (%s %s);\n", table.Name, idCol.Name, idCol.Type)
			_, err := db1.Exec(fmt.Sprintf("CREATE TABLE %s (%s %s);", table.Name, idCol.Name, idCol.Type))
			if err != nil {
				return err
			}
		}

		// add columns
		for _, column := range table.Columns {
			if column.Name == "id" {
				continue
			}
			// TODO: same issue as above
			log.Printf("sql: ALTER TABLE %s ADD COLUMN %s %s;\n", table.Name, column.Name, column.Type)
			_, err := db1.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, column.Name, column.Type))
			if err != nil {
				return err
			}
		}

		fmt.Printf("Added table %s with %d columns.\n", table.Name, len(table.Columns))
	}
	fmt.Printf("Done adding %d tables.\n", len(migration.Add))

	return nil
}
