package resolver

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestDatabaseIDResolveResponseFirstID(t *testing.T) {
	tests := []struct {
		name string
		resp databaseIDResolveResponse
		want string
	}{
		{
			name: "databaseId",
			resp: databaseIDResolveResponse{DatabaseID: " db-1 "},
			want: "db-1",
		},
		{
			name: "id",
			resp: databaseIDResolveResponse{ID: " id-1 "},
			want: "id-1",
		},
		{
			name: "database.databaseId",
			resp: databaseIDResolveResponse{
				Database: struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{DatabaseID: " db-2 "},
			},
			want: "db-2",
		},
		{
			name: "database.id",
			resp: databaseIDResolveResponse{
				Database: struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{ID: " id-2 "},
			},
			want: "id-2",
		},
		{
			name: "databases.databaseId",
			resp: databaseIDResolveResponse{
				Databases: []struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{{DatabaseID: " db-3 "}},
			},
			want: "db-3",
		},
		{
			name: "databases.id",
			resp: databaseIDResolveResponse{
				Databases: []struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{{ID: " id-3 "}},
			},
			want: "id-3",
		},
		{
			name: "data.databaseId",
			resp: databaseIDResolveResponse{
				Data: []struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{{DatabaseID: " db-4 "}},
			},
			want: "db-4",
		},
		{
			name: "data.id",
			resp: databaseIDResolveResponse{
				Data: []struct {
					DatabaseID string `json:"databaseId"`
					ID         string `json:"id"`
				}{{ID: " id-4 "}},
			},
			want: "id-4",
		},
		{
			name: "empty",
			resp: databaseIDResolveResponse{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.firstID(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestExtractDatabaseIDFromAPIKey(t *testing.T) {
	if got := extractDatabaseIDFromAPIKey(" "); got != "" {
		t.Fatalf("expected empty from blank key, got %q", got)
	}
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	if got := extractDatabaseIDFromAPIKey("key_" + uuid); got != uuid {
		t.Fatalf("expected uuid %q, got %q", uuid, got)
	}
	if got := extractDatabaseIDFromAPIKey("prefix db_abc123 suffix"); got != "db_abc123" {
		t.Fatalf("expected db token, got %q", got)
	}
	if got := extractDatabaseIDFromAPIKey("key_plain"); got != "" {
		t.Fatalf("expected empty for non-matching key, got %q", got)
	}
}

func TestResolveDatabaseIDFromAPIKey(t *testing.T) {
	ctx := context.Background()
	origLookup := lookupDatabaseIDFromAPIKeyFn
	t.Cleanup(func() { lookupDatabaseIDFromAPIKeyFn = origLookup })

	lookupDatabaseIDFromAPIKeyFn = func(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
		t.Fatalf("unexpected lookup call")
		return "", nil
	}
	uuid := "123e4567-e89b-12d3-a456-426614174000"
	if got, err := resolveDatabaseIDFromAPIKey(ctx, "http://base", "key_"+uuid, "secret"); err != nil || got != uuid {
		t.Fatalf("expected derived uuid, got %q err %v", got, err)
	}

	if got, err := resolveDatabaseIDFromAPIKey(ctx, "http://base", "", "secret"); err != nil || got != "" {
		t.Fatalf("expected empty for missing key, got %q err %v", got, err)
	}
	if got, err := resolveDatabaseIDFromAPIKey(ctx, "http://base", "key", ""); err != nil || got != "" {
		t.Fatalf("expected empty for missing secret, got %q err %v", got, err)
	}

	if _, err := resolveDatabaseIDFromAPIKey(ctx, "", "key", "secret"); err == nil {
		t.Fatalf("expected error for missing base url")
	}

	lookupDatabaseIDFromAPIKeyFn = func(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
		if baseURL != "http://base" || apiKey != "key_plain" || apiSecret != "secret" {
			t.Fatalf("unexpected lookup args: %s %s %s", baseURL, apiKey, apiSecret)
		}
		return "db-from-lookup", nil
	}
	if got, err := resolveDatabaseIDFromAPIKey(ctx, "http://base", "key_plain", "secret"); err != nil || got != "db-from-lookup" {
		t.Fatalf("expected lookup result, got %q err %v", got, err)
	}
}

func TestLookupDatabaseIDFromAPIKeySuccessOnFirstPath(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first", "/second"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	calledSecond := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/first":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"databaseId":"db-first"}`))
		case "/second":
			calledSecond = true
			http.Error(w, `{"code":"oops","message":"should-not"}`, http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	got, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret")
	if err != nil {
		t.Fatalf("lookup err: %v", err)
	}
	if got != "db-first" {
		t.Fatalf("expected db-first, got %s", got)
	}
	if calledSecond {
		t.Fatalf("expected second path to be skipped")
	}
}

func TestLookupDatabaseIDFromAPIKeyFallbackToSecondPath(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first", "/second"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/first":
			http.Error(w, `{"code":"not_found","message":"missing"}`, http.StatusNotFound)
		case "/second":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"database":{"id":"db-second"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	got, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret")
	if err != nil {
		t.Fatalf("lookup err: %v", err)
	}
	if got != "db-second" {
		t.Fatalf("expected db-second, got %s", got)
	}
}

func TestLookupDatabaseIDFromAPIKeyDebugLogging(t *testing.T) {
	t.Setenv("ONYX_DEBUG", "true")
	origWriter := log.Writer()
	origFlags := log.Flags()
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(origWriter)
		log.SetFlags(origFlags)
	})

	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first", "/second"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/first":
			http.Error(w, `{"code":"not_found","message":"missing"}`, http.StatusNotFound)
		case "/second":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"databaseId":"db-second"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	got, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret")
	if err != nil {
		t.Fatalf("lookup err: %v", err)
	}
	if got != "db-second" {
		t.Fatalf("expected db-second, got %s", got)
	}

	out := buf.String()
	if !strings.Contains(out, "resolve database id GET") {
		t.Fatalf("expected debug log for GET, got %q", out)
	}
	if !strings.Contains(out, "failed for") {
		t.Fatalf("expected debug log for failure, got %q", out)
	}
}

func TestLookupDatabaseIDFromAPIKeyNonRetryableError(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first", "/second"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	calledSecond := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/first":
			http.Error(w, `{"code":"boom","message":"fail"}`, http.StatusInternalServerError)
		case "/second":
			calledSecond = true
			http.Error(w, `{"code":"oops","message":"should-not"}`, http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	if _, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret"); err == nil {
		t.Fatalf("expected error")
	}
	if calledSecond {
		t.Fatalf("expected second path to be skipped")
	}
}

func TestLookupDatabaseIDFromAPIKeyMissingDatabaseID(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	if _, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret"); err == nil {
		t.Fatalf("expected error for missing databaseId")
	}
}

func TestLookupDatabaseIDFromAPIKeyAllRetryableErrors(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = []string{"/first", "/second"}
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"not_found","message":"missing"}`, http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	if _, err := lookupDatabaseIDFromAPIKey(context.Background(), srv.URL, "key", "secret"); err == nil {
		t.Fatalf("expected error for retryable failures")
	}
}

func TestLookupDatabaseIDFromAPIKeyNoPaths(t *testing.T) {
	origPaths := databaseIDResolvePaths
	databaseIDResolvePaths = nil
	t.Cleanup(func() { databaseIDResolvePaths = origPaths })

	if _, err := lookupDatabaseIDFromAPIKey(context.Background(), "http://example.com", "key", "secret"); err == nil {
		t.Fatalf("expected error with no resolve paths")
	}
}

func TestShouldTryNextResolvePath(t *testing.T) {
	if shouldTryNextResolvePath(errors.New("nope")) {
		t.Fatalf("expected false for non-contract error")
	}
	if !shouldTryNextResolvePath(&contract.Error{Meta: map[string]any{"status": 404}}) {
		t.Fatalf("expected true for 404")
	}
	if !shouldTryNextResolvePath(&contract.Error{Meta: map[string]any{"status": float64(405)}}) {
		t.Fatalf("expected true for 405 float")
	}
	if shouldTryNextResolvePath(&contract.Error{Meta: map[string]any{"status": 500}}) {
		t.Fatalf("expected false for 500")
	}
}

func TestResolveDatabaseIDDeriveError(t *testing.T) {
	isolateResolverEnv(t)
	ClearCache()
	orig := resolveDatabaseIDFromAPIKeyFn
	resolveDatabaseIDFromAPIKeyFn = func(ctx context.Context, baseURL, apiKey, apiSecret string) (string, error) {
		return "", errors.New("boom")
	}
	t.Cleanup(func() { resolveDatabaseIDFromAPIKeyFn = orig })

	if _, _, err := Resolve(context.Background(), Config{APIKey: "key", APISecret: "secret"}); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected derive error, got %v", err)
	}
}
