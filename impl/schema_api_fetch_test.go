package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestFetchSchemaSupportsHistoryAndNestedSchema(t *testing.T) {
	// First call returns history array, second returns nested schema object.
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		w.Header().Set("Content-Type", "application/json")
		if call == 1 {
			w.Write([]byte(`{"schemas":[{"entities":[{"name":"A","attributes":[{"name":"id","type":"String"}]}]}]}`))
			return
		}
		w.Write([]byte(`{"schema":{"tables":[{"name":"B","fields":[{"name":"id","type":"String"}]}]}}`))
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db_test"},
	}

	schema, err := fetchSchema(context.Background(), c, []string{"A"})
	if err != nil {
		t.Fatalf("fetchSchema err: %v", err)
	}
	if _, ok := schema.Table("A"); !ok {
		t.Fatalf("expected table from history response")
	}

	schema, err = fetchSchema(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchSchema nested err: %v", err)
	}
	if _, ok := schema.Table("B"); !ok {
		t.Fatalf("expected table from nested schema object")
	}
}
