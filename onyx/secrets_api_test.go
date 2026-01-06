package onyx

import (
	"context"
	"net/http"
	"testing"
)

func TestSecretsAPI(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/secrets":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"key":"API_KEY","value":"abc"}]`))
		case "/secrets/API_KEY":
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

	secrets, err := c.Secrets().List(context.Background())
	if err != nil || len(secrets) != 1 {
		t.Fatalf("list err: %v secrets: %+v", err, secrets)
	}

	secret, err := c.Secrets().Get(context.Background(), "API_KEY")
	if err != nil || secret.Key != "API_KEY" {
		t.Fatalf("get err: %v secret: %+v", err, secret)
	}

	saved, err := c.Secrets().Set(context.Background(), secret)
	if err != nil || saved.UpdatedAt == "" {
		t.Fatalf("set err: %v saved: %+v", err, saved)
	}

	if err := c.Secrets().Delete(context.Background(), "API_KEY"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
}
