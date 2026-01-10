package main

import (
	"os"

	"github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func main() {
	exit(runMain(os.Args[1:]))
}

var exit = os.Exit

func runMain(args []string) int {
	return commands.Dispatch(args)
}
