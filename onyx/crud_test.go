package onyx

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSaveDelete(t *testing.T) {
	saved := false
	deleted := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if r.URL.Path != "/tables/users/save" {
				t.Fatalf("unexpected save path: %s", r.URL.Path)
			}
			saved = true
		case http.MethodDelete:
			if r.URL.Path != "/tables/users/123" {
				t.Fatalf("unexpected delete path: %s", r.URL.Path)
			}
			deleted = true
		}
	})

	if err := c.Save(context.Background(), "users", map[string]any{"id": "123"}); err != nil {
		t.Fatalf("save err: %v", err)
	}
	if err := c.Delete(context.Background(), "users", "123"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
	if !saved || !deleted {
		t.Fatalf("expected operations to be invoked")
	}
}

func TestBatchSaveChunking(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
	})

	entities := []any{"a", "b", "c"}
	if err := c.BatchSave(context.Background(), "users", entities, 2); err != nil {
		t.Fatalf("batch err: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestBatchSaveRetry(t *testing.T) {
	attempts := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	if err := c.BatchSave(context.Background(), "users", []any{"a"}, 1); err != nil {
		t.Fatalf("batch err: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected retry, got %d attempts", attempts)
	}
}

func TestCascadeOperations(t *testing.T) {
	saved := false
	deleted := false
	cascadeSpec := contract.Cascade("graph:type(field)")

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			saved = true
		}
		if r.Method == http.MethodDelete {
			deleted = true
		}
	})

	cc := c.Cascade(cascadeSpec)
	if err := cc.Save(context.Background(), "users", map[string]any{}); err != nil {
		t.Fatalf("cascade save err: %v", err)
	}
	if err := cc.Delete(context.Background(), "users", "id"); err != nil {
		t.Fatalf("cascade delete err: %v", err)
	}
	if !saved || !deleted {
		t.Fatalf("expected cascade operations")
	}
}
