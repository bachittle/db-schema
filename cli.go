package main

import (
	"fmt"
	"io"
	"os"
	"sort"
)

type cli_context struct {
	descr      string
	executable string
	commands   map[string]cmd
}

func (cc *cli_context) root_usage(w io.Writer) {
	max_n := 8
	names := []string{}
	for c := range cc.commands {
		names = append(names, c)
		n := len(c)
		if n > max_n {
			max_n = n
		}
	}
	sort.Strings(names)

	fmt.Fprintf(w, "usage: %s <command> [args]\n\n", cc.executable)
	fmt.Fprintln(w, "the commands are:")
	for _, cmd := range names {
		fmt.Fprintf(w, "    %-*s  %s\n", max_n, cmd, cc.commands[cmd].short_descr())
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "use '%s help <command>' for more information about a command.\n", cc.executable)
}

func (cc *cli_context) execute(args []string) error {
	helping := false
	if len(args) > 0 && args[0] == "help" {
		helping = true
		args = args[1:]
	}

	if len(args) == 0 {
		if !helping {
			fmt.Fprintf(os.Stderr, "a command is required\n\n")
			cc.root_usage(os.Stderr)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "%s %s\n\n", cc.executable, cc.descr)
		cc.root_usage(os.Stdout)
		os.Exit(0)
	}

	cmdname := args[0]
	args = args[1:]
	c, ok := cc.commands[cmdname]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %s\n\n", cmdname)
		fmt.Fprintf(os.Stderr, "run '%s help'.\n", cc.executable)
		os.Exit(1)
	}

	if helping {
		fmt.Fprintf(os.Stdout, "%s\n\n", c.short_descr())
		fmt.Fprintf(os.Stdout, "usage: %s %s %s\n\n", cc.executable, cmdname, c.args())
		c.long_descr(os.Stdout)
		fmt.Fprintf(os.Stdout, "\n")
		os.Exit(0)
	}

	err := c.execute(args)
	if err == nil {
		return nil
	}

	if v, ok := err.(*ErrInvalidArgs); ok {
		fmt.Fprintf(os.Stderr, "error in %s: %s\n\n", cmdname, v.Wrapped.Error())
		fmt.Fprintf(os.Stdout, "usage: %s %s %s\n\n", cc.executable, cmdname, c.args())
		fmt.Fprintf(os.Stdout, "use '%s help %s' for more information.\n", cc.executable, cmdname)
		os.Exit(1)
	}

	return err
}

type cmd interface {
	args() string
	short_descr() string
	long_descr(out io.Writer)
	execute(args []string) error
}

type ErrInvalidArgs struct {
	Wrapped error
}

func (e ErrInvalidArgs) Error() string {
	return e.Wrapped.Error()
}
