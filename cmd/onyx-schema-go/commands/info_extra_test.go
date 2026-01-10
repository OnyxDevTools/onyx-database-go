package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
)

func TestDefaultConnectInfoClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/database/db/schema" {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tables":[]}`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	cfg := resolver.ResolvedConfig{
		DatabaseID:      "db",
		DatabaseBaseURL: srv.URL,
		APIKey:          "k",
		APISecret:       "s",
	}

	client, err := defaultConnectInfoClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("defaultConnectInfoClient err: %v", err)
	}
	if _, err := client.Schema(context.Background()); err != nil {
		t.Fatalf("schema fetch err: %v", err)
	}
}
