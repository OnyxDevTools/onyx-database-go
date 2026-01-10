package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDiffCommandLoadError(t *testing.T) {
	tmp := t.TempDir()
	pathA := filepath.Join(tmp, "a.json")
	_ = os.WriteFile(pathA, []byte("{"), 0o644)

	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &DiffCommand{}
	if code := cmd.Run([]string{"--a", pathA, "--b", filepath.Join(tmp, "missing.json")}); code != 1 {
		t.Fatalf("expected failure, got %d", code)
	}
}
