package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adnsv/go-db3/schema"
	"gopkg.in/yaml.v3"
)

type scan_cmd struct {
	input_fn   string
	output_fn  string
	output_fmt string
	normtypes  bool
	normnames  string
	sort       string
}

func (v *scan_cmd) flags() *flag.FlagSet {
	ff := flag.NewFlagSet("", flag.ContinueOnError)
	ff.StringVar(&v.output_fn, "output", "", "output filename")
	ff.StringVar(&v.output_fmt, "fmt", "", "output data format (json|yaml|sql)")
	ff.StringVar(&v.sort, "sort", "", "sort output tables/columns/indices")
	ff.StringVar(&v.sort, "normnames", "", "normalize names tables/columns/indices to upper or lower case")
	ff.BoolVar(&v.normtypes, "normtypes", false, "normalize column types, remove default nulls for nullables")
	return ff
}
func (v *scan_cmd) args() string {
	return "[-output=<filename>] [-normtypes] [-normnames=<...>] [-sort=<...>] DBFILE"
}
func (v *scan_cmd) short_descr() string { return "retrieve schema from database" }
func (v *scan_cmd) long_descr(out io.Writer) {

	ff := v.flags()
	ff.SetOutput(out)
	ff.PrintDefaults()

	fmt.Fprint(out, `
Scans SQLite database file for existing schema.

Specify output filename with the '-output=<filename>' flag, 
otherwise the output is dumped to stdout.

The extracted schema can be dumped in json, yaml, or sql formats 
(guessed from the output file extension, defaults to json 
for console output).

Override format with '-fmt=json' '-fmt=yaml' or '-fmt=sql' flags.

With the '-normtypes' flag, column types can be normalized:

	int integer tinyint smallint mediumint   -> int
	int64 bigint                             -> int64
	boolean bool                             -> bool
	real double float                        -> float
	blob                                     -> blob
	text string clob                         -> text
	date                                     -> date
	time                                     -> time
	datetime timestamp                       -> timestamp
	uuid, guid                               -> uuid
	character() varchar() nchar() nvarchar() -> text

Notice, that '-normtypes' also removed 'default null' from nullable
columns (that don't have 'not null' in their type).

With the '-normnames=<upper|lower>' flag, the names for all
the tables, columns, and indices can be converted to upper or lower
case.

With the '-sort' flag, tables/columns/indices can be sorted:

  -sort=tables                 sorts tables by name
  -sort=columns,indices        sorts content within each table
  -sort=tables,columns,indices sorts everything (nice for diffing)

Note: sorting is performed lexicographically. If '-normnames' is 
also specified, sorting is performed after all the name normalization.
`)

}
func (v *scan_cmd) execute(args []string) error {
	ff := v.flags()
	err := ff.Parse(args)
	if err != nil {
		return &ErrInvalidArgs{Wrapped: err}
	}

	aa := ff.Args()
	if len(aa) == 1 {
		v.input_fn = aa[0]
	} else {
		return &ErrInvalidArgs{Wrapped: fmt.Errorf("missing DBFILE parameter")}
	}

	// validate output path
	if v.output_fn != "" {
		stat, err := os.Stat(v.output_fn)
		if err == nil && stat.IsDir() {
			return fmt.Errorf("invalid output file name '%s': resolves to an existing directory (expected path to a file within an existing directory)", v.output_fn)
		}
		output_dir := filepath.Dir(v.output_fn)
		stat, err = os.Stat(output_dir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("invalid output file name '%s': parent directory does not exist", v.output_fn)
			}
			return fmt.Errorf("invalid output filename: %w", err)
		} else if !stat.IsDir() {
			return fmt.Errorf("invalid output file name '%s': expected path to a file within an existing directory)", v.output_fn)
		}
	}

	if v.output_fmt == "" {
		if v.output_fn != "" {
			// try to guess output format from output filename extension
			ext := filepath.Ext(v.output_fn)
			switch strings.ToLower(ext) {
			case ".json":
				v.output_fmt = "json"
			case ".yaml", ".yml":
				v.output_fmt = "yaml"
			case ".sql":
				v.output_fmt = "sql"
			case "":
				return &ErrInvalidArgs{Wrapped: fmt.Errorf("output file name does not have an extension, please specify output format: -fmt=json|yaml")}
			default:
				return &ErrInvalidArgs{Wrapped: fmt.Errorf("unknown output file name extension '%s', please specify output format: -fmt=json|yaml", ext)}
			}
		} else {
			// if no output filename is specified, default to json
			v.output_fmt = "json"
		}
	} else {
		switch v.output_fmt {
		case "json":
		case "sql":
		case "yaml", "yml":
			v.output_fmt = "yaml"
		default:
			return &ErrInvalidArgs{Wrapped: fmt.Errorf("unsupported output format %s", v.output_fmt)}
		}
	}

	conn, err := sql.Open("sqlite3", v.input_fn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	sch, err := schema.Scan(conn)
	if err != nil {
		log.Fatal(err)
	}

	if v.normtypes {
		for _, t := range sch.Tables {
			for _, f := range t.Columns {
				f.Type = schema.NormalizeType(f.Type)
				schema.NormalizeDefault(f)
			}
		}
	}
	if v.normnames != "" {
		var uppercase bool
		switch v.normnames {
		case "upper":
			uppercase = true
		case "lower":
			uppercase = false
		default:
			return &ErrInvalidArgs{Wrapped: fmt.Errorf("unsupported normnames value %s", v.normnames)}
		}
		for _, t := range sch.Tables {
			schema.NormalizeNames(t, uppercase)
		}
	}
	if v.sort != "" {
		what := "," + v.sort + ","
		do_tables := strings.Contains(what, ",tables,")
		do_columns := strings.Contains(what, ",columns,")
		do_indices := strings.Contains(what, ",indices,")
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

	switch v.output_fmt {
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

	if v.output_fn == "" {
		_, err = os.Stdout.Write(out)
	} else {
		fmt.Fprintf(os.Stderr, "writing results to %s ... ", v.output_fn)
		err = os.WriteFile(v.output_fn, out, 0666)
		if err == nil {
			fmt.Fprintf(os.Stderr, "SUCCEEDED\n")
		} else {
			fmt.Fprintf(os.Stderr, "FAILED\n")
		}
	}
	return err
}
