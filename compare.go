package main

import (
	"fmt"
	"os"

	"github.com/bachittle/go-db3/schema"
	"gopkg.in/yaml.v3"
)

type CompareCmd struct {
	BaseSchemaFile    string `arg:"" type:"existingfile" help:"The base schema file (the version that may need to be migrated)."`
	DerivedSchemaFile string `arg:"" type:"existingfile" help:"The derived schema file (the latest version, reference point for base schema)."`
}

func (v *CompareCmd) Run() error {
	fmt.Println("# compare", v.BaseSchemaFile, v.DerivedSchemaFile)

	// load base and derived
	baseData, err := os.ReadFile(v.BaseSchemaFile)
	if err != nil {
		return err
	}
	derivedData, err := os.ReadFile(v.DerivedSchemaFile)
	if err != nil {
		return err
	}

	// convert to yaml
	var baseYaml, derivedYaml []*schema.Table
	err = yaml.Unmarshal(baseData, &baseYaml)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(derivedData, &derivedYaml)
	if err != nil {
		return err
	}

	// compare
	migration := compareSchemas(baseYaml, derivedYaml)
	err = yaml.NewEncoder(os.Stdout).Encode(migration)

	return err
}

// Migration contains the instructions to migrate a database from one schema to another
// base => add items, remove items => derived
type Migration struct {
	Add SubMigration `yaml:"add"`
	Rm  SubMigration `yaml:"rm"`
}

// SubMigration is a set of instructions to add or remove a specific type of item
// e.g. add entire table, or add column to existing table
// insertion updates if the item already exists, otherwise it creates a new item
type SubMigration []*schema.Table

// Append adds a schema to the submigration.
// If the table already exists, it appends the columns to the existing table,
// otherwise it creates a new table.
func (s *SubMigration) Append(sch *schema.Table) {
	for _, existingSchema := range *s {
		if existingSchema.Name == sch.Name {
			for _, column := range sch.Columns {
				existingSchema.Columns = append(existingSchema.Columns, column)
			}
			return
		}
	}

	// table doesn't exist, add it
	*s = append(*s, sch)
}

// compareSchemas compares two schemas and returns migration instructions
func compareSchemas(base, derived []*schema.Table) Migration {
	var migration Migration

	compareTables(base, derived, &migration)
	compareColumns(base, derived, &migration)

	return migration
}

// compareTables compares base and derived tables and adds migration instructions
func compareTables(base, derived []*schema.Table, migration *Migration) {
	// add tables
	for _, derivedSchema := range derived {
		found := false
		for _, baseSchema := range base {
			if derivedSchema.Name == baseSchema.Name {
				found = true
				break
			}
		}
		if !found {
			migration.Add.Append(derivedSchema)
		}
	}

	// remove tables
	for _, baseSchema := range base {
		found := false
		for _, derivedSchema := range derived {
			if derivedSchema.Name == baseSchema.Name {
				found = true
				break
			}
		}
		if !found {
			migration.Rm.Append(baseSchema)
		}
	}
}

// compareColumns compares base and derived columns and adds migration instructions
func compareColumns(base, derived []*schema.Table, migration *Migration) {
	// add columns
	for _, derivedSchema := range derived {
		for _, baseSchema := range base {
			if derivedSchema.Name != baseSchema.Name {
				continue
			}
			for _, derivedColumn := range derivedSchema.Columns {
				found := false
				for _, baseColumn := range baseSchema.Columns {
					if derivedColumn.Name == baseColumn.Name {
						found = true
						break
					}
				}
				if !found {
					// TODO: make this more efficient by appending all columns at the end
					migration.Add.Append(&schema.Table{
						Name: derivedSchema.Name,
						Columns: []*schema.Column{
							derivedColumn,
						},
					})
				}
			}
		}
	}
}
