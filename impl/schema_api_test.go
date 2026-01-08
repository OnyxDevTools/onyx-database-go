package impl

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSchemaFetch(t *testing.T) {
	schemaJSON := `{"tables":{"users":{"fields":{"id":{"type":"string"}}}}}`
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(schemaJSON))
	})

	schema, err := c.Schema(context.Background())
	if err != nil {
		t.Fatalf("schema err: %v", err)
	}
	if _, ok := schema.Table("users"); !ok {
		t.Fatalf("schema not parsed: %+v", schema)
	}
}

func TestSchemaPublish(t *testing.T) {
	called := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/schemas/db_test" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.PublishSchema(context.Background(), contract.Schema{Tables: []contract.Table{{
		Name:   "B",
		Fields: []contract.Field{{Name: "z"}},
	}}})
	if err != nil {
		t.Fatalf("publish err: %v", err)
	}
	if !called {
		t.Fatalf("expected publish request")
	}
}
