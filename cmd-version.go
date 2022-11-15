package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

var app_ver string = ""

func app_version() string {
	v, ok := debug.ReadBuildInfo()
	if ok && v.Main.Version != "(devel)" {
		// installed with go install
		return v.Main.Version
	} else if app_ver != "" {
		// built with ld-flags
		return app_ver
	} else {
		return "#UNAVAILABLE"
	}
}

type version_cmd struct{}

func (v *version_cmd) args() string        { return "" }
func (v *version_cmd) short_descr() string { return "prints application version information" }
func (v *version_cmd) long_descr(out io.Writer) {
	fmt.Fprintln(out, "Prints the application version.")
}
func (v *version_cmd) execute(args []string) error {
	fmt.Fprintln(os.Stdout, app_version())
	return nil
}
