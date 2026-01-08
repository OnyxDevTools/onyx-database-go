package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// Stdout and Stderr allow commands to direct output; tests can override.
var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

// Command represents a CLI subcommand.
type Command interface {
	Name() string
	Description() string
	Run(args []string) int
}

// Dispatch runs the appropriate subcommand based on the provided args.
// Exit codes: 0 success, 1 failure, 2 usage error.
func Dispatch(args []string) int {
	cmds := availableCommands()
	if len(args) == 0 {
		printRootUsage(cmds)
		return 2
	}

	if args[0] == "-h" || args[0] == "--help" {
		printRootUsage(cmds)
		return 0
	}

	registry := map[string]Command{}
	for _, c := range cmds {
		registry[c.Name()] = c
	}

	cmd, ok := registry[args[0]]
	if !ok {
		fmt.Fprintf(Stderr, "unknown command %q\n", args[0])
		printRootUsage(cmds)
		return 2
	}

	return cmd.Run(args[1:])
}

func printRootUsage(cmds []Command) {
	fmt.Fprintln(Stdout, "Usage: onyx-schema-go <command> [options]")
	fmt.Fprintln(Stdout)
	fmt.Fprintln(Stdout, "Available commands:")

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name() < cmds[j].Name()
	})

	for _, c := range cmds {
		fmt.Fprintf(Stdout, "  %-10s %s\n", c.Name(), c.Description())
	}
}

var availableCommands = defaultAvailableCommands

func defaultAvailableCommands() []Command {
	return []Command{
		&ValidateCommand{},
		&DiffCommand{},
		&NormalizeCommand{},
		&GetCommand{},
		&PublishCommand{},
	}
}
