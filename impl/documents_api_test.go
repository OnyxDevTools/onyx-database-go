package impl

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDocumentsAPI(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/db_test/document":
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodPut {
				if _, err := w.Write([]byte(`{"documentId":"doc_1","data":{"foo":"bar"},"updatedAt":"now"}`)); err != nil {
					t.Fatalf("write response: %v", err)
				}
				return
			}
			if _, err := w.Write([]byte(`[{"documentId":"doc_1","data":{"foo":"bar"}}]`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
		case "/data/db_test/document/doc_1":
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"documentId":"doc_1","data":{"foo":"bar"}}`)); err != nil {
					t.Fatalf("write response: %v", err)
				}
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

func TestNormalizeDocumentIDs(t *testing.T) {
	doc := contract.OnyxDocument{ID: "x"}
	normalized := normalizeDocumentIDs(doc)
	if normalized.DocumentID != "x" {
		t.Fatalf("expected documentId populated, got %+v", normalized)
	}
}

func TestDocumentClientValidationAndPreferredID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"documentId":"doc_pref"}`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
		}
	})

	// Missing IDs
	if _, err := c.Documents().Save(context.Background(), contract.OnyxDocument{}); err == nil {
		t.Fatalf("expected error on missing id")
	}
	if _, err := c.Documents().Get(context.Background(), " "); err == nil {
		t.Fatalf("expected error on empty id")
	}
	if err := c.Documents().Delete(context.Background(), ""); err == nil {
		t.Fatalf("expected error on empty id for delete")
	}

	doc, err := c.Documents().Save(context.Background(), contract.OnyxDocument{DocumentID: "doc_pref", ID: "ignored"})
	if err != nil {
		t.Fatalf("save err: %v", err)
	}
	if doc.DocumentID != "doc_pref" || doc.ID != "doc_pref" {
		t.Fatalf("preferred id not applied: %+v", doc)
	}

	// Ensure fallback uses ID when DocumentID is empty
	c2 := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"id":"only_id"}`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
		}
	})
	saved, err := c2.Documents().Save(context.Background(), contract.OnyxDocument{ID: "only_id"})
	if err != nil || saved.DocumentID != "only_id" {
		t.Fatalf("expected fallback id applied: err=%v saved=%+v", err, saved)
	}

	c3 := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"documentId":"trimmed"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	})
	docTrim, err := c3.Documents().Save(context.Background(), contract.OnyxDocument{DocumentID: "  trimmed  "})
	if err != nil || docTrim.DocumentID != "trimmed" {
		t.Fatalf("expected trimmed document id, got %+v err=%v", docTrim, err)
	}
}
