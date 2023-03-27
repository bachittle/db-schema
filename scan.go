package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bachittle/go-db3/schema"
	"gopkg.in/yaml.v3"

	_ "github.com/mattn/go-sqlite3"
)

type ScanCmd struct {
	Output    string   `short:"o" type:"path" help:"Output filename."`
	Fmt       string   `short:"f" enum:"yaml,json,sql" default:"yaml" placeholder:"yaml|json|sql" help:"Output format."`
	NormTypes bool     `optional:"" help:"Normalize column types, remove default nulls for nullables."`
	NormNames string   `enum:"skip,upper,lower" default:"skip" placeholder:"upper|lower" help:"Normalize names to upper/lower case."`
	Sort      []string `short:"s" optional:"" enum:"tables|columns|indices" placeholder:"PART" help:"Sort output (tables/columns/indices)."`

	InputFile string `arg:"" required:"" type:"existingfile" help:"Path to SQLite database file."`
}

func (v *ScanCmd) Run(ctx *kong.Context) error {
	// validate output path
	if v.Output != "" {
		stat, err := os.Stat(v.Output)
		if err == nil && stat.IsDir() {
			return fmt.Errorf("invalid output file name '%s': resolves to an existing directory (expected path to a file within an existing directory)", v.Output)
		}
		output_dir := filepath.Dir(v.Output)
		stat, err = os.Stat(output_dir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("invalid output file name '%s': parent directory does not exist", v.Output)
			}
			return fmt.Errorf("invalid output filename: %w", err)
		} else if !stat.IsDir() {
			return fmt.Errorf("invalid output file name '%s': expected path to a file within an existing directory)", v.Output)
		}
	}

	if v.Fmt == "" {
		if v.Output != "" {
			// try to guess output format from output filename extension
			ext := filepath.Ext(v.Output)
			switch strings.ToLower(ext) {
			case ".json":
				v.Fmt = "json"
			case ".yaml", ".yml":
				v.Fmt = "yaml"
			case ".sql":
				v.Fmt = "sql"
			case "":
				return fmt.Errorf("output file name does not have an extension, please specify output format: --fmt=json|yaml|sql")
			default:
				return fmt.Errorf("unknown output file name extension '%s', please specify output format: --fmt=json|yaml|sql", ext)
			}
		} else {
			// if no output filename is specified, default to json
			v.Fmt = "json"
		}
	} else {
		switch v.Fmt {
		case "json":
		case "sql":
		case "yaml", "yml":
			v.Fmt = "yaml"
		default:
			return fmt.Errorf("unsupported output format %s", v.Fmt)
		}
	}

	conn, err := sql.Open("sqlite3", v.InputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	sch, err := schema.Scan(conn)
	if err != nil {
		log.Fatal(err)
	}

	if v.NormTypes {
		for _, t := range sch.Tables {
			for _, f := range t.Columns {
				f.Type = schema.NormalizeType(f.Type)
				schema.NormalizeDefault(f)
			}
		}
	}
	if v.NormNames != "skip" {
		var uppercase bool
		switch v.NormNames {
		case "upper":
			uppercase = true
		case "lower":
			uppercase = false
		default:
			return fmt.Errorf("unsupported norm-names value %s", v.NormNames)
		}
		for _, t := range sch.Tables {
			schema.NormalizeNames(t, uppercase)
		}
	}
	if len(v.Sort) > 0 {
		m := map[string]bool{}
		for _, v := range v.Sort {
			m[v] = true
		}
		do_tables := m["tables"]
		do_columns := m["columns"]
		do_indices := m["indices"]
		if do_tables {
			schema.SortTables(sch.Tables)
		}
		if do_columns || do_indices {
			for _, t := range sch.Tables {
				if do_columns {
					schema.SortColumns(t)
				}
				if do_indices {
					schema.SortIndices(t)
				}
			}
		}
	}

	var out []byte

	switch v.Fmt {
	case "json":
		out, err = json.MarshalIndent(sch.Tables, "", "    ")
	case "yaml":
		out, err = yaml.Marshal(sch.Tables)
	case "sql":
		b := bytes.Buffer{}
		for i, t := range sch.Tables {
			if i > 0 {
				b.WriteByte('\n')
			}
			t.CreateStatements(&b)
		}
		out = b.Bytes()
	}
	if err != nil {
		return err
	}

	if v.Output == "" {
		_, err = os.Stdout.Write(out)
	} else {
		fmt.Fprintf(os.Stderr, "writing results to %s ... ", v.Output)
		err = os.WriteFile(v.Output, out, 0666)
		if err == nil {
			fmt.Fprintf(os.Stderr, "SUCCEEDED\n")
		} else {
			fmt.Fprintf(os.Stderr, "FAILED\n")
		}
	}
	return err
}
