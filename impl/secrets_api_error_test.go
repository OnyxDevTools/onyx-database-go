package impl

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestSecretsAPIErrorPaths(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})

	if _, err := c.Secrets().Get(context.Background(), "k"); err == nil {
		t.Fatalf("expected get error")
	}
	if _, err := c.Secrets().Set(context.Background(), contract.OnyxSecret{Key: "k"}); err == nil {
		t.Fatalf("expected set error")
	}
	if _, err := c.Secrets().List(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}
}
