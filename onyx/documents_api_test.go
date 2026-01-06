package onyx

import (
	"context"
	"net/http"
	"testing"
)

func TestDocumentsAPI(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/documents":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id":"doc_1","data":{"foo":"bar"}}]`))
		case "/documents/doc_1":
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"id":"doc_1","data":{"foo":"bar"}}`))
			} else if r.Method == http.MethodPut {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"id":"doc_1","data":{"foo":"bar"},"updatedAt":"now"}`))
			} else if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	docs, err := c.Documents().List(context.Background())
	if err != nil || len(docs) != 1 {
		t.Fatalf("list err: %v docs: %+v", err, docs)
	}

	doc, err := c.Documents().Get(context.Background(), "doc_1")
	if err != nil || doc.ID != "doc_1" {
		t.Fatalf("get err: %v doc: %+v", err, doc)
	}

	saved, err := c.Documents().Save(context.Background(), doc)
	if err != nil || saved.UpdatedAt == "" {
		t.Fatalf("save err: %v saved: %+v", err, saved)
	}

	if err := c.Documents().Delete(context.Background(), "doc_1"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}
