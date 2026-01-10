package impl

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamIteratorValue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, `{"id":1}`+"\n"); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("get err: %v", err)
	}
	iter := newStreamIterator(resp)
	if !iter.Next() {
		t.Fatalf("expected item")
	}
	if val := iter.Value(); val == nil || val["id"] != float64(1) {
		t.Fatalf("unexpected value: %+v", val)
	}
	iter.Close()
}

func TestStreamIteratorHandlesEmptyAndErrors(t *testing.T) {
	// Empty body returns false without error
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(""))}
	iter := newStreamIterator(resp)
	if iter.Next() {
		t.Fatalf("expected no data")
	}
	if iter.Err() != nil {
		t.Fatalf("expected nil error, got %v", iter.Err())
	}
	iter.Close()

	// Pre-set error short-circuits
	resp2 := &http.Response{Body: io.NopCloser(strings.NewReader("{}"))}
	iter2 := newStreamIterator(resp2).(*streamIterator)
	iter2.err = io.EOF
	if iter2.Next() {
		t.Fatalf("expected false when err preset")
	}
}
