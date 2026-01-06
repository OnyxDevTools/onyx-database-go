package onyx

import (
	"context"
	"net/http"
	"testing"
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
