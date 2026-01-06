package onyx

import (
	"context"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestMetricsRequiresIdentifier(t *testing.T) {
	c := &client{}
	if _, err := c.Metrics(context.Background(), contract.MetricsParams{}); err == nil {
		t.Fatalf("expected error when uid and email are empty")
	}
}

func TestMetricsQueryParameters(t *testing.T) {
	var seenUID, seenEmail string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		seenUID = r.URL.Query().Get("uid")
		seenEmail = r.URL.Query().Get("email")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"requests":3}`))
	})

	metrics, err := c.Metrics(context.Background(), contract.MetricsParams{UID: "user-1", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("metrics err: %v", err)
	}

	if seenUID != "user-1" || seenEmail != "user@example.com" {
		t.Fatalf("unexpected query params: uid=%s email=%s", seenUID, seenEmail)
	}

	if val, ok := metrics["requests"]; !ok || val != float64(3) {
		t.Fatalf("unexpected metrics payload: %+v", metrics)
	}
}
