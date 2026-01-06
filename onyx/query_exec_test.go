package onyx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	hc := httpclient.New(srv.URL, srv.Client(), httpclient.Options{})
	return &client{httpClient: hc}
}

func TestQueryList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/query/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1}]`))
	})

	q := &query{client: c, table: "users"}
	res, err := q.List(context.Background())
	if err != nil {
		t.Fatalf("list err: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("unexpected results: %+v", res)
	}
}

func TestQueryPage(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"items":[{"id":1}],"nextCursor":"abc"}`))
	})

	q := &query{client: c, table: "users"}
	res, err := q.Page(context.Background(), "")
	if err != nil {
		t.Fatalf("page err: %v", err)
	}
	if res.NextCursor != "abc" {
		t.Fatalf("unexpected cursor: %+v", res)
	}
}

func TestQueryStream(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}\n{"))
	})

	q := &query{client: c, table: "users"}
	it, err := q.Stream(context.Background())
	if err != nil {
		t.Fatalf("stream err: %v", err)
	}
	defer it.Close()
	if !it.Next() {
		t.Fatalf("expected first element")
	}
	if it.Err() != nil {
		t.Fatalf("unexpected error: %v", it.Err())
	}
	// scanner will error on malformed second line
	it.Next()
	if it.Err() == nil {
		t.Fatalf("expected error")
	}
}
