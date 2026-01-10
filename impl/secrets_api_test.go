package impl

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSecretsAPI(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/database/db_test/secret":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"records":[{"key":"API_KEY","value":"abc"}]}`))
		case "/database/db_test/secret/API_KEY":
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"key":"API_KEY","value":"abc"}`))
			} else if r.Method == http.MethodPut {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"key":"API_KEY","value":"abc","updatedAt":"now"}`))
			} else if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	secrets, err := c.ListSecrets(context.Background())
	if err != nil || len(secrets) != 1 {
		t.Fatalf("list err: %v secrets: %+v", err, secrets)
	}

	secret, err := c.GetSecret(context.Background(), "API_KEY")
	if err != nil || secret.Key != "API_KEY" {
		t.Fatalf("get err: %v secret: %+v", err, secret)
	}

	saved, err := c.PutSecret(context.Background(), secret)
	if err != nil || saved.UpdatedAt == "" {
		t.Fatalf("set err: %v saved: %+v", err, saved)
	}

	if err := c.DeleteSecret(context.Background(), "API_KEY"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}

func TestSecretsAPIValidationErrors(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := c.GetSecret(context.Background(), ""); err == nil {
		t.Fatalf("expected get validation error")
	}
	if _, err := c.PutSecret(context.Background(), contract.OnyxSecret{}); err == nil {
		t.Fatalf("expected put validation error")
	}
	if err := c.DeleteSecret(context.Background(), ""); err == nil {
		t.Fatalf("expected delete validation error")
	}
}

func TestSecretsAPIErrorPropagation(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"nope"}`, http.StatusBadRequest)
	})

	if _, err := c.Secrets().List(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}
}

func TestOnyxSecretAlias(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"records":[]}`))
	})
	if _, err := c.OnyxSecret().List(context.Background()); err != nil {
		t.Fatalf("expected alias to delegate, got %v", err)
	}
	if calls == 0 {
		t.Fatalf("expected underlying call to occur")
	}
}
