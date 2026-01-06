package commands

import "fmt"

// StubCommand is a placeholder for future functionality.
type StubCommand struct {
	name        string
	description string
}

func (c *StubCommand) Name() string        { return c.name }
func (c *StubCommand) Description() string { return c.description }

func (c *StubCommand) Run(args []string) int {
	fmt.Fprintf(Stdout, "%s command is not yet implemented.\n", c.name)
	return 1
}
