package impl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	hc := httpclient.New(srv.URL, srv.Client(), httpclient.Options{})
	return &client{httpClient: hc, cfg: resolver.ResolvedConfig{DatabaseID: "db_test"}}
}

func TestQueryList(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/db_test/query/users" {
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

func TestQueryUpdate(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/db_test/query/update/users" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`1`))
	})

	q := (&query{client: c, table: "users"}).SetUpdates(map[string]any{"email": "x"})
	updated, err := q.Update(context.Background())
	if err != nil {
		t.Fatalf("update err: %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected 1 row updated, got %d", updated)
	}
}

func TestQueryDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/db_test/query/delete/users" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`1`))
	})

	q := &query{client: c, table: "users"}
	deleted, err := q.Delete(context.Background())
	if err != nil {
		t.Fatalf("delete err: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("unexpected deleted count: %d", deleted)
	}
}

func TestQueryPageWithParamsAndError(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			raw := r.URL.RawQuery
			if !(strings.Contains(raw, "pageSize=3") && strings.Contains(raw, "nextPage=next")) {
				t.Fatalf("unexpected query params: %s", raw)
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"code":"bad","message":"fail"}`))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"items":[{"id":2}]}`))
		}
	})

	limit := 3
	q := &query{client: c, table: "users", limit: &limit}
	if _, err := q.Page(context.Background(), "next"); err == nil {
		t.Fatalf("expected error on first attempt")
	}

	res, err := q.Page(context.Background(), "next")
	if err != nil || len(res.Items) != 1 {
		t.Fatalf("retry page err: %v res: %+v", err, res)
	}
}

func TestQueryOperationsErrorPaths(t *testing.T) {
	errClient := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})

	q := &query{client: errClient, table: "users"}
	if _, err := q.List(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}
	if _, err := q.Update(context.Background()); err == nil {
		t.Fatalf("expected update error")
	}
	if _, err := q.Delete(context.Background()); err == nil {
		t.Fatalf("expected delete error")
	}
	if _, err := q.Stream(context.Background()); err == nil {
		t.Fatalf("expected stream error")
	}
}
