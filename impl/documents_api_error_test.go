package impl

import (
	"context"
	"net/http"
	"testing"
)

func TestDocumentsAPIErrorPaths(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})

	if _, err := c.Documents().List(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}
	if _, err := c.Documents().Get(context.Background(), "id"); err == nil {
		t.Fatalf("expected get error")
	}
	if err := c.Documents().Delete(context.Background(), "id"); err == nil {
		t.Fatalf("expected delete error")
	}
}

func TestDocumentsListNilItems(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`null`))
		}
	})
	if docs, err := c.Documents().List(context.Background()); err != nil || docs != nil {
		t.Fatalf("expected nil slice without error, got docs=%v err=%v", docs, err)
	}
}
