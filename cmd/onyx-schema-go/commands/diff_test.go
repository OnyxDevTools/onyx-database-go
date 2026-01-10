package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
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

	if !strings.Contains(out.String(), "Tables only in updated schema") {
		t.Fatalf("expected human-readable output, got: %s", out.String())
	}

	header := "Comparing schemas (base=" + pathA + ", updated=" + pathB + " (file))"
	if !strings.Contains(out.String(), header) {
		t.Fatalf("expected header with sources, got: %s", out.String())
	}
}

func TestDiffCommandFetchesFromAPI(t *testing.T) {
	tmpDir := t.TempDir()
	pathA := filepath.Join(tmpDir, "a.json")
	writeFile(t, pathA, `{"tables": []}`)

	original := fetchSchemaFromAPI
	defer func() { fetchSchemaFromAPI = original }()

	var called bool
	fetchSchemaFromAPI = func(ctx context.Context, databaseID string) (onyx.Schema, error) {
		called = true
		if databaseID != "db-123" {
			t.Fatalf("expected database id db-123, got %s", databaseID)
		}
		return onyx.Schema{
			Tables: []onyx.Table{{Name: "users", Fields: []onyx.Field{{Name: "id", Type: "string"}}}},
		}, nil
	}

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &DiffCommand{}
	exitCode := cmd.Run([]string{"--a", pathA, "--database-id", "db-123", "--json"})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d (out: %s)", exitCode, out.String())
	}

	if !called {
		t.Fatal("expected fetchSchemaFromAPI to be called")
	}

	if !strings.Contains(out.String(), `"removedTables"`) {
		t.Fatalf("expected json diff output, got: %s", out.String())
	}
}

func TestDiffCommandDefaultsToAPIWhenBMissing(t *testing.T) {
	tmpDir := t.TempDir()
	pathA := filepath.Join(tmpDir, "a.json")
	writeFile(t, pathA, `{"tables": []}`)

	original := fetchSchemaFromAPI
	defer func() { fetchSchemaFromAPI = original }()

	var called bool
	fetchSchemaFromAPI = func(ctx context.Context, databaseID string) (onyx.Schema, error) {
		if databaseID != "" {
			t.Fatalf("expected empty database id, got %s", databaseID)
		}
		called = true
		return onyx.Schema{
			Tables: []onyx.Table{{Name: "users", Fields: []onyx.Field{{Name: "id", Type: "string"}}}},
		}, nil
	}

	var out bytes.Buffer
	Stdout = &out
	Stderr = &out
	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
	}()

	cmd := &DiffCommand{}
	exitCode := cmd.Run([]string{"--a", pathA})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d (out: %s)", exitCode, out.String())
	}

	if !called {
		t.Fatal("expected fetchSchemaFromAPI to be called")
	}

	header := "Comparing schemas (base=API (configured credentials), updated=" + pathA + " (file))"
	if !strings.Contains(out.String(), header) {
		t.Fatalf("expected header showing API source, got: %s", out.String())
	}
}
