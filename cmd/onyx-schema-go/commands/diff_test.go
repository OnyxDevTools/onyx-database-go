package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiffCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	pathA := filepath.Join(tmpDir, "a.json")
	pathB := filepath.Join(tmpDir, "b.json")

	writeFile(t, pathA, `{
  "tables": [
    {"name": "users", "fields": [
      {"name": "id", "type": "string", "nullable": false}
    ]}
  ]
}`)

	writeFile(t, pathB, `{
  "tables": [
    {"name": "users", "fields": [
      {"name": "id", "type": "string", "nullable": true},
      {"name": "email", "type": "string"}
    ]},
    {"name": "events", "fields": []}
  ]
}`)

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &DiffCommand{}
	exitCode := cmd.Run([]string{"--a", pathA, "--b", pathB, "--json"})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	got := out.String()
	expected := `{
  "addedTables": [
    {
      "name": "events",
      "fields": []
    }
  ],
  "tableDiffs": [
    {
      "name": "users",
      "addedFields": [
        {
          "name": "email",
          "type": "string"
        }
      ],
      "modifiedFields": [
        {
          "name": "id",
          "from": {
            "name": "id",
            "type": "string"
          },
          "to": {
            "name": "id",
            "type": "string",
            "nullable": true
          }
        }
      ]
    }
  ]
}
`

	if got != expected {
		t.Fatalf("unexpected diff output. got:\n%s\nexpected:\n%s", got, expected)
	}
}

func TestDiffCommandHumanReadable(t *testing.T) {
	tmpDir := t.TempDir()
	pathA := filepath.Join(tmpDir, "a.json")
	pathB := filepath.Join(tmpDir, "b.json")

	writeFile(t, pathA, `{"tables": []}`)
	writeFile(t, pathB, `{"tables": [{"name": "users", "fields": []}]}`)

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &DiffCommand{}
	exitCode := cmd.Run([]string{"--a", pathA, "--b", pathB})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}

	if !strings.Contains(out.String(), "Added tables") {
		t.Fatalf("expected human-readable output, got: %s", out.String())
	}
}

func TestDiffCommandRequiresB(t *testing.T) {
	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &DiffCommand{}
	exitCode := cmd.Run([]string{})
	if exitCode != 2 {
		t.Fatalf("expected usage exit 2, got %d", exitCode)
	}
}
