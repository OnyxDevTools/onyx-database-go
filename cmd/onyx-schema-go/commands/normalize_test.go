package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCommandWritesDeterministicOutput(t *testing.T) {
	tmpDir := t.TempDir()
	input := filepath.Join(tmpDir, "schema.json")
	output := filepath.Join(tmpDir, "normalized.json")

	writeFile(t, input, `{
  "tables": [
    {"name": "b", "fields": [{"name": "b_field", "type": "string"}]},
    {"name": "a", "fields": [{"name": "a_field", "type": "string"}]}
  ]
}`)

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &NormalizeCommand{}
	exitCode := cmd.Run([]string{"--schema", input, "--out", output})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	first, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read normalized output: %v", err)
	}

	out.Reset()
	exitCode = cmd.Run([]string{"--schema", output, "--out", output})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	second, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read normalized output: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("normalizing twice should produce identical bytes")
	}
}
