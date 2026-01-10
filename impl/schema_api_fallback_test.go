package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestFetchSchemaFallbacks(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		switch call {
		case 1: // initial /schema -> 404 to trigger fallback chain
			http.Error(w, `{"code":"not_found","message":"missing"}`, http.StatusNotFound)
		case 2: // /schemas also fails to force legacy /schema
			http.Error(w, `{"code":"nope","message":"fail"}`, http.StatusInternalServerError)
		default: // legacy /schema success with tables map
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"tables":{"users":{"fields":{"id":{"type":"string"}}}}}`)); err != nil {
				t.Fatalf("write response: %v", err)
			}
		}
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db_test"},
	}

	schema, err := fetchSchema(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchSchema err: %v", err)
	}
	if _, ok := schema.Table("users"); !ok {
		t.Fatalf("expected users table from fallback response")
	}
	if call < 3 {
		t.Fatalf("expected multiple fallbacks, got %d calls", call)
	}
}

func TestStripEntityTextAndHelpers(t *testing.T) {
	raw := map[string]any{
		"entityText": "remove me",
		"nested": []any{
			map[string]any{"entityText": "inner", "keep": "ok"},
		},
	}
	cleaned := stripEntityText(raw).(map[string]any)
	if _, ok := cleaned["entityText"]; ok {
		t.Fatalf("expected entityText removed at root")
	}
	if inner, ok := cleaned["nested"].([]any); !ok || inner[0].(map[string]any)["entityText"] != nil {
		t.Fatalf("expected nested entityText removed")
	}

	if stringValue(nil) != "" {
		t.Fatalf("stringValue default mismatch")
	}
	if boolValue("nope") {
		t.Fatalf("boolValue default mismatch")
	}
	if mapValue(123) != nil {
		t.Fatalf("mapValue default mismatch")
	}
}
