package impl

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStreamIteratorErrorOnInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "not-json\n"); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	iter := newStreamIterator(resp)
	if iter.Next() {
		t.Fatalf("expected Next to fail on invalid json")
	}
	if iter.Err() == nil {
		t.Fatalf("expected error")
	}
}
