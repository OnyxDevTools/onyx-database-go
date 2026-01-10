package impl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestBatchSaveUsesDefaultBatchSize(t *testing.T) {
	callCount := 0
	var lastCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var payload []map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		lastCount = len(payload)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
		sleep:      func(time.Duration) {},
	}

	entities := make([]any, 0, defaultBatchSize+1)
	for i := 0; i < defaultBatchSize+1; i++ {
		entities = append(entities, map[string]any{"id": i})
	}

	if err := batchSave(context.Background(), c, "users", entities, 0); err != nil {
		t.Fatalf("batch save err: %v", err)
	}
	if callCount != 2 || lastCount != 1 {
		t.Fatalf("expected batching into two calls, got %d calls lastCount=%d", callCount, lastCount)
	}
}
