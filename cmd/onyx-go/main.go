package main

import (
	"fmt"
	"io"
	"os"

	schemaCmds "github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func main() {
	exit(runMain(os.Args, os.Stdout, os.Stderr))
}

var exit = os.Exit

func runMain(args []string, stdout, stderr io.Writer) int {
	return dispatch(args[1:], stdout, stderr)
}

func dispatch(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stdout)
		return 2
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootUsage(stdout)
		return 0
	case "schema":
		schemaCmds.Stdout = stdout
		schemaCmds.Stderr = stderr
		return schemaCmds.Dispatch(args[1:])
	default:
		fmt.Fprintf(stderr, "unknown subcommand %q\n", args[0])
		printRootUsage(stderr)
		return 2
	}
}

func printRootUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: onyx-go <subcommand> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  schema    Schema operations (validate/diff/get/publish)")
}
