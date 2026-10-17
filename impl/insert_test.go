package impl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/msgpack"
)

func TestInsertPreservesWireFormatAndConflict(t *testing.T) {
	for _, format := range []string{contract.WireFormatJSON, contract.WireFormatMessagePack} {
		t.Run(format, func(t *testing.T) {
			calls := 0
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/data/db_test/Observation" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if calls == 2 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(`{"error":{"message":"Entity already exists"}}`))
					return
				}
				var value any
				if format == contract.WireFormatMessagePack {
					value = requireMessagePackRequest(t, r)
					w.Header().Set("Content-Type", msgpack.MediaType)
					w.WriteHeader(http.StatusCreated)
					writeMessagePack(t, w, value)
				} else {
					if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
						t.Fatal(err)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(value)
				}
			})
			c.wireFormat = format
			row := map[string]any{"id": "new", "value": 2}
			inserted, err := c.Insert(context.Background(), "Observation", row, nil)
			if err != nil || inserted["id"] != "new" {
				t.Fatalf("inserted=%v err=%v", inserted, err)
			}
			_, err = c.Insert(context.Background(), "Observation", row, nil)
			var apiError *contract.Error
			if !errors.As(err, &apiError) {
				t.Fatalf("expected contract error, got %v", err)
			}
			if apiError.Message != "Entity already exists" {
				t.Fatalf("message=%q", apiError.Message)
			}
			if apiError.Meta["status"] != http.StatusConflict {
				t.Fatalf("status=%v", apiError.Meta["status"])
			}
			if calls != 2 {
				t.Fatalf("calls=%d; insert must not retry", calls)
			}
		})
	}
}
