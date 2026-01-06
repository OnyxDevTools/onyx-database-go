package main

import (
	"os"

	"github.com/OnyxDevTools/onyx-database-go/cmd/onyx-schema-go/commands"
)

func main() {
	os.Exit(commands.Dispatch(os.Args[1:]))
}
