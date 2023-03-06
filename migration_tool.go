// the migration tool keeps track of schemas with its own sqlite database
// cmd frontend gives tools to interact with migrations (create, apply, rollback, etc)
package DbSchema

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adnsv/go-db3/schema"
)

type MigrationToolCmd struct {
	Create   MigrationCreateCmd   `cmd:"" help:"Create a migration and store it in migration database."`
	Upgrade  MigrationUpgradeCmd  `cmd:"" help:"Perform a migration upgrade on a potentially out-of-date database."`
	Rollback MigrationRollbackCmd `cmd:"" help:"Perform a migration rollback to previous version."`
}

type MigrationCreateCmd struct {
	Host    string `arg:"" type:"existingfile" help:"The database host to base the migration on."`
	DbStore string `arg:"" optional:"" type:"file" help:"The sqlite database that the migration tool uses to keep track of migrations."`
}

type MigrationUpgradeCmd struct {
	Host    string `arg:"" type:"existingfile" help:"The database host to perform a migration upgrade on."`
	DbStore string `arg:"" type:"existingfile" help:"The sqlite database that the migration tool uses to keep track of migrations."`
}

type MigrationRollbackCmd struct {
	Host    string `arg:"" type:"existingfile" help:"The database host to perform a migration rollback on."`
	DbStore string `arg:"" type:"existingfile" help:"The sqlite database that the migration tool uses to keep track of migrations."`
}

func (v *MigrationCreateCmd) Run() error {
	if v.DbStore == "" {
		v.DbStore = v.Host + ".migrations.db"
	}
	fmt.Println("creating migration database")
	fmt.Println("host:", v.Host)
	fmt.Println("migrations db:", v.DbStore)

	// we can open an sql connection without writing a file
	// files are only created when we execute a query
	migratorDB, err := sql.Open("sqlite3", v.DbStore)
	if err != nil {
		return err
	}
	defer migratorDB.Close()

	// if file does not exist, ask for confirmation to create it
	if _, err := os.Stat(v.DbStore); os.IsNotExist(err) {
		fmt.Print("The migration database does not exist. Create it? [y/n]: ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input)[0] != 'y' {
			fmt.Println("Aborting.")
			return nil
		}
		fmt.Println("Creating schema...")

		err = initMigrationsDb(migratorDB)
		if err != nil {
			return err
		}
	}

	// load schema from host
	hostDB, err := sql.Open("sqlite3", v.Host)
	if err != nil {
		return err
	}
	defer hostDB.Close()

	// a host db exists, we can create a migration

	var description string
	fmt.Print("Enter a description for the migration: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		description = scanner.Text()
	}

	// ask for confirmation
	fmt.Println("Creating migration for", filepath.Base(v.Host), "with description:", description)
	fmt.Print("Insert new migration? [y/n]: ")
	var input string
	fmt.Scanln(&input)
	if strings.ToLower(input)[0] != 'y' {
		fmt.Println("Exiting cleanly.")
		return nil
	}

	_, err = migratorDB.Exec(`INSERT INTO migrations (name, description) VALUES (?, ?)`, filepath.Base(v.Host), description)
	if err != nil {
		return err
	}

	sch, err := schema.Scan(hostDB)
	// _, err = schema.Scan(hostDB)
	if err != nil {
		return err
	}
	fmt.Println("schema loaded:")

	for _, table := range sch.Tables {
		fmt.Println(table.Name)
	}

	return nil
}

func initMigrationsDb(db *sql.DB) error {
	// create migrations table
	_, err := db.Exec(`CREATE TABLE migrations (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	return nil
}
