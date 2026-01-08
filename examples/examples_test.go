package examples_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExamples(t *testing.T) {
	if os.Getenv("ONYX_RUN_EXAMPLES") == "" {
		t.Skip("set ONYX_RUN_EXAMPLES=1 to run example binaries")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Dir(wd)

	examples := []string{
		"./examples/cmd/seed",
		"./examples/delete/cmd/byid",
		"./examples/delete/cmd/query",
		"./examples/document/cmd/savegetdelete",
		"./examples/query/cmd/aggregateavg",
		"./examples/query/cmd/aggregateswithgrouping",
		"./examples/query/cmd/basic",
		"./examples/query/cmd/compound",
		"./examples/query/cmd/findbyid",
		"./examples/query/cmd/firstornull",
		"./examples/query/cmd/innerquery",
		"./examples/query/cmd/inpartition",
		"./examples/query/cmd/list",
		"./examples/query/cmd/notinnerquery",
		"./examples/query/cmd/orderby",
		"./examples/query/cmd/resolver",
		"./examples/query/cmd/searchbyresolverfields",
		"./examples/query/cmd/select",
		"./examples/query/cmd/sortingandpaging",
		"./examples/query/cmd/update",
		"./examples/save/cmd/basic",
		"./examples/save/cmd/batchsave",
		"./examples/save/cmd/cascade",
		"./examples/save/cmd/cascadebuilder",
		"./examples/schema/cmd/basic",
		"./examples/secrets/cmd/basic",
		"./examples/stream/cmd/close",
		"./examples/stream/cmd/createevents",
		"./examples/stream/cmd/deleteevents",
		"./examples/stream/cmd/querystream",
		"./examples/stream/cmd/updateevents",
	}

	for _, path := range examples {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "go", "run", path)
			cmd.Dir = repoRoot
			cmd.Env = os.Environ()

			var output bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &output

			if err := cmd.Run(); err != nil {
				t.Fatalf("run %s: %v\noutput:\n%s", path, err, output.String())
			}

			if !strings.Contains(output.String(), "example: completed") {
				t.Fatalf("expected completion log in %s output:\n%s", path, output.String())
			}
		})
	}
}
