package commands

import (
	"bytes"
	"os"
	"testing"
)

type fakeCommand struct {
	name        string
	description string
	code        int
	args        []string
}

func (f *fakeCommand) Name() string        { return f.name }
func (f *fakeCommand) Description() string { return f.description }
func (f *fakeCommand) Run(args []string) int {
	f.args = args
	return f.code
}

func TestDispatchUsageAndUnknown(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	if code := Dispatch(nil); code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}

	if code := Dispatch([]string{"--help"}); code != 0 {
		t.Fatalf("expected help exit code 0, got %d", code)
	}

	if code := Dispatch([]string{"does-not-exist"}); code != 2 {
		t.Fatalf("expected unknown exit code 2, got %d", code)
	}
}

func TestDispatchInvokesCommand(t *testing.T) {
	Stdout, Stderr = &bytes.Buffer{}, &bytes.Buffer{}
	defer func() { Stdout, Stderr = os.Stdout, os.Stderr }()

	cmd := &fakeCommand{name: "fake", description: "desc", code: 0}
	availableCommands = func() []Command { return []Command{cmd} }
	defer func() { availableCommands = defaultAvailableCommands }()

	if code := Dispatch([]string{"fake", "arg1"}); code != 0 {
		t.Fatalf("expected command exit 0, got %d", code)
	}
	if len(cmd.args) != 1 || cmd.args[0] != "arg1" {
		t.Fatalf("expected args to be forwarded, got %+v", cmd.args)
	}
}
