package impl

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func newSchemaTestClient(t *testing.T, handler http.HandlerFunc) *client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	hc := httpclient.New(srv.URL, srv.Client(), httpclient.Options{})
	return &client{httpClient: hc, cfg: resolver.ResolvedConfig{DatabaseID: "db_test"}}
}

func TestClientSchemaEndpoints(t *testing.T) {
	c := newSchemaTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/schemas/db_test/validate":
			w.WriteHeader(http.StatusOK)
		case "/schemas/db_test/history":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]contract.Schema{{Tables: []contract.Table{{Name: "T"}}}})
		case "/schemas/db_test":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	schema := contract.Schema{Tables: []contract.Table{{Name: "User"}}}
	if err := c.UpdateSchema(context.Background(), schema, true); err != nil {
		t.Fatalf("UpdateSchema err: %v", err)
	}
	if err := c.ValidateSchema(context.Background(), schema); err != nil {
		t.Fatalf("ValidateSchema err: %v", err)
	}
	history, err := c.GetSchemaHistory(context.Background())
	if err != nil || len(history) != 1 {
		t.Fatalf("GetSchemaHistory err: %v hist: %+v", err, history)
	}
}

func TestClientFromConstructsQuery(t *testing.T) {
	c := &client{cfg: resolver.ResolvedConfig{DatabaseID: "db"}}
	q := c.From("users")
	if q == nil {
		t.Fatalf("expected query")
	}
}

func TestInitUsesDebugLoggingAndCacheKey(t *testing.T) {
	t.Setenv("ONYX_DATABASE_ID", "db")
	t.Setenv("ONYX_DATABASE_BASE_URL", "http://example.com")
	t.Setenv("ONYX_DATABASE_API_KEY", "k")
	t.Setenv("ONYX_DATABASE_API_SECRET", "s")
	t.Setenv("ONYX_DEBUG", "true")
	resolver.ClearCache()

	c, err := Init(context.Background(), contract.Config{})
	if err != nil {
		t.Fatalf("init err: %v", err)
	}
	implClient := c.(*client)
	signer := httpclient.Signer{APIKey: "k", APISecret: "s"}
	key := httpClientCacheKey("http://example.com", nil, true, true, signer)
	if key == "" {
		t.Fatalf("expected cache key")
	}
	logger1 := log.New(io.Discard, "", 0)
	cached := getCachedHTTPClient("http://example.com", nil, true, true, signer, logger1)
	if cached != implClient.httpClient {
		t.Fatalf("expected cached http client reuse")
	}
	if implClient.sleep == nil {
		t.Fatalf("sleep should be initialized")
	}

	customHTTP := &http.Client{}
	keyWithHTTP := httpClientCacheKey("http://example.com", customHTTP, true, true, signer)
	if keyWithHTTP == "" {
		t.Fatalf("expected cache key with base http client")
	}
	logger2 := log.New(io.Discard, "", 0)
	second := getCachedHTTPClient("http://example.com", customHTTP, true, true, signer, logger2)
	if second == nil {
		t.Fatalf("expected cached client returned")
	}
}
