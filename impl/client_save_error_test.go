package impl

import (
	"context"
	"net/http"
	"testing"
)

func TestClientSaveCascadeSpecNil(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {})
	if _, err := c.Save(context.Background(), "users", map[string]any{}, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := c.Save(context.Background(), "users", map[string]any{}, []string{""})
	if err != nil {
		t.Fatalf("did not expect error for empty cascade spec: %v", err)
	}
}
