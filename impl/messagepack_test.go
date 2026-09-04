package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/msgpack"
)

func writeMessagePack(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	body, err := msgpack.Marshal(value)
	if err != nil {
		t.Fatalf("encode MessagePack response: %v", err)
	}
	w.Header().Set("Content-Type", msgpack.MediaType)
	if _, err := w.Write(body); err != nil {
		t.Fatalf("write MessagePack response: %v", err)
	}
}

func requireMessagePackRequest(t *testing.T, r *http.Request) any {
	t.Helper()
	if got := r.Header.Get("Accept"); got != msgpack.MediaType+", application/json;q=0.9" {
		t.Fatalf("Accept = %q", got)
	}
	if r.Body == nil || r.ContentLength == 0 {
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Fatalf("bodyless Content-Type = %q, want empty", got)
		}
		return nil
	}
	if got := r.Header.Get("Content-Type"); got != msgpack.MediaType {
		t.Fatalf("Content-Type = %q, want %q", got, msgpack.MediaType)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	var value any
	if err := msgpack.Unmarshal(body, &value); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	return value
}

func TestMessagePackEntityCRUDAndQuery(t *testing.T) {
	const largeID = int64(1<<53 + 1)
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		request := requireMessagePackRequest(t, r)
		switch r.URL.Path {
		case "/data/db_test/users":
			if r.Method == http.MethodDelete {
				t.Fatal("entity delete should include an id path segment")
			}
			entity, ok := request.(map[string]any)
			if !ok || entity["name"] != "Ada" {
				t.Fatalf("save request = %#v", request)
			}
			writeMessagePack(t, w, map[string]any{"id": int64(7), "name": "Ada"})
		case "/data/db_test/users/7":
			if r.Method != http.MethodDelete || request != nil {
				t.Fatalf("delete method=%s request=%#v", r.Method, request)
			}
			w.WriteHeader(http.StatusNoContent)
		case "/data/db_test/query/users":
			query, ok := request.(map[string]any)
			if !ok || query["type"] != "SelectQuery" || query["table"] != "users" {
				t.Fatalf("query request = %#v", request)
			}
			if r.URL.Query().Get("nextPage") == "cursor-1" {
				writeMessagePack(t, w, map[string]any{
					"records":  []any{map[string]any{"id": largeID + 2}},
					"nextPage": "cursor-2",
				})
				return
			}
			writeMessagePack(t, w, map[string]any{
				"records": []any{map[string]any{
					"id":   largeID,
					"name": "Ada",
					"nested": map[string]any{
						"revision": largeID + 1,
					},
				}},
			})
		case "/data/db_test/query/update/users":
			query, ok := request.(map[string]any)
			if !ok || query["type"] != "UpdateQuery" {
				t.Fatalf("update request = %#v", request)
			}
			writeMessagePack(t, w, int64(2))
		case "/data/db_test/query/delete/users":
			writeMessagePack(t, w, int64(3))
		default:
			t.Fatalf("unexpected entity path %s", r.URL.Path)
		}
	})
	c.wireFormat = contract.WireFormatMessagePack

	saved, err := c.Save(context.Background(), "users", map[string]any{"name": "Ada"}, nil)
	if err != nil || saved["id"] != int64(7) {
		t.Fatalf("Save result=%#v err=%v", saved, err)
	}

	q := &query{client: c, table: "users"}
	rows, err := q.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0]["id"] != largeID {
		t.Fatalf("List result=%#v err=%v", rows, err)
	}
	nested, ok := rows[0]["nested"].(map[string]any)
	if !ok || nested["revision"] != largeID+1 {
		t.Fatalf("List nested signed integer was not preserved: %#v", rows)
	}

	page, err := q.Page(context.Background(), "cursor-1")
	if err != nil || page.NextCursor != "cursor-2" || len(page.Items) != 1 {
		t.Fatalf("Page result=%#v err=%v", page, err)
	}
	if page.Items[0]["id"] != largeID+2 {
		t.Fatalf("Page signed integer was not preserved: %#v", page)
	}

	updated, err := q.SetUpdates(map[string]any{"active": true}).Update(context.Background())
	if err != nil || updated != 2 {
		t.Fatalf("Update result=%d err=%v", updated, err)
	}

	deleted, err := q.Delete(context.Background())
	if err != nil || deleted != 3 {
		t.Fatalf("query Delete result=%d err=%v", deleted, err)
	}
	if err := c.Delete(context.Background(), "users", "7"); err != nil {
		t.Fatalf("entity Delete: %v", err)
	}
}

func TestMessagePackBatchSave(t *testing.T) {
	var batchSizes []int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		request := requireMessagePackRequest(t, r)
		batch, ok := request.([]any)
		if !ok {
			t.Fatalf("batch request = %T %#v", request, request)
		}
		batchSizes = append(batchSizes, len(batch))
		w.WriteHeader(http.StatusNoContent)
	})
	c.wireFormat = contract.WireFormatMessagePack

	entities := []any{
		map[string]any{"id": int64(1)},
		map[string]any{"id": int64(2)},
		map[string]any{"id": int64(3)},
	}
	if err := c.BatchSave(context.Background(), "users", entities, 2); err != nil {
		t.Fatalf("BatchSave: %v", err)
	}
	if len(batchSizes) != 2 || batchSizes[0] != 2 || batchSizes[1] != 1 {
		t.Fatalf("batch sizes = %v", batchSizes)
	}
}

func TestMessagePackConcatenatedEntityStream(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = requireMessagePackRequest(t, r)
		w.Header().Set("Content-Type", msgpack.MediaType)
		for _, value := range []any{
			nil,
			map[string]any{"id": int64(1)},
			map[string]any{"id": int64(2)},
		} {
			body, err := msgpack.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(body); err != nil {
				t.Fatal(err)
			}
		}
	})
	c.wireFormat = contract.WireFormatMessagePack

	it, err := (&query{client: c, table: "users"}).Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer it.Close()
	for want := int64(1); want <= 2; want++ {
		if !it.Next() || it.Value()["id"] != want {
			t.Fatalf("stream value=%#v next=%v err=%v", it.Value(), false, it.Err())
		}
	}
	if it.Next() || it.Err() != nil {
		t.Fatalf("stream completion Next=%v Err=%v", false, it.Err())
	}
}

func TestNonEntityRoutesRemainJSONWithMessagePackConfigured(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type for %s = %q", r.URL.Path, got)
		}
		if strings.HasSuffix(r.URL.Path, "/document") {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode JSON document: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"documentId":"doc-1"}`))
			return
		}
		t.Fatalf("unexpected path %s", r.URL.Path)
	})
	c.wireFormat = contract.WireFormatMessagePack

	doc, err := c.Documents().Save(context.Background(), contract.OnyxDocument{DocumentID: "doc-1"})
	if err != nil || doc.DocumentID != "doc-1" {
		t.Fatalf("document result=%#v err=%v", doc, err)
	}
}

func TestWireFormatNormalization(t *testing.T) {
	for input, want := range map[string]string{
		"":          contract.WireFormatMessagePack,
		"  JSON  ":  contract.WireFormatJSON,
		" MSGPACK ": contract.WireFormatMessagePack,
	} {
		got, err := normalizeWireFormat(input)
		if err != nil || got != want {
			t.Fatalf("normalizeWireFormat(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := normalizeWireFormat("protobuf"); err == nil {
		t.Fatal("expected unsupported wire format error")
	}
	if _, err := Init(context.Background(), Config{WireFormat: "protobuf"}); err == nil {
		t.Fatal("Init should reject an unsupported wire format before resolving config")
	}
}

func TestMessagePackStreamRejectsNonObjectFrame(t *testing.T) {
	body, err := msgpack.Marshal(int64(1))
	if err != nil {
		t.Fatal(err)
	}
	resp := &http.Response{
		Header: http.Header{"Content-Type": []string{msgpack.MediaType}},
		Body:   io.NopCloser(bytes.NewReader(body)),
	}
	it := newStreamIterator(resp)
	if it.Next() || it.Err() == nil {
		t.Fatalf("expected invalid stream frame error, got %v", it.Err())
	}
}
