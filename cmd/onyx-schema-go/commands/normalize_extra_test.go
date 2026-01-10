package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCommandWriteSchema(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "schema.json")
	if err := os.WriteFile(src, []byte(`{"tables":[{"name":"users","fields":[{"name":"id","type":"string"}]}]}`), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	var out bytes.Buffer
	Stdout, Stderr = &out, &out
	t.Cleanup(func() { Stdout, Stderr = os.Stdout, os.Stderr })

	cmd := &NormalizeCommand{}
	code := cmd.Run([]string{"--schema", src})
	if code != 0 {
		t.Fatalf("expected success, got %d (%s)", code, out.String())
	}
}
