package commands

import (
	"bytes"
	"os"
	"testing"
)

func TestStubCommand(t *testing.T) {
	Stdout = &bytes.Buffer{}
	defer func() { Stdout = os.Stdout }()

	cmd := &StubCommand{name: "stub", description: "desc"}
	if cmd.Name() != "stub" || cmd.Description() != "desc" {
		t.Fatalf("unexpected name/description")
	}
	if code := cmd.Run(nil); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
