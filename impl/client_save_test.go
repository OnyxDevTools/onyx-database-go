package impl

import (
	"context"
	"net/http"
	"testing"
)

func TestClientSaveWithRelationships(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "relationships=roles,perms" && r.URL.RawQuery != "relationships=roles%2Cperms" {
			t.Fatalf("expected relationships query, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"id":"1"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})

	resp, err := c.Save(context.Background(), "users", map[string]any{"id": "1"}, []string{"roles", "perms"})
	if err != nil || resp["id"] != "1" {
		t.Fatalf("save err: %v resp:%+v", err, resp)
	}
}

func TestGetSchemaHistoryError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})
	if _, err := c.GetSchemaHistory(context.Background()); err == nil {
		t.Fatalf("expected error from history fetch")
	}
}

func TestGetSchemaViaFacade(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"tables":[{"name":"Users","fields":[{"name":"id","type":"string"}]}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})
	if _, err := c.Schema(context.Background()); err != nil {
		t.Fatalf("schema facade err: %v", err)
	}
}
