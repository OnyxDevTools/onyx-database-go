package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/internal/msgpack"
)

func TestDoEntityMessagePack(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != msgpack.MediaType {
			t.Errorf("Content-Type = %q", got)
		}
		if got := r.Header.Get("Accept"); got != msgpack.MediaType+", application/json;q=0.9" {
			t.Errorf("Accept = %q", got)
		}
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			return
		}
		var request map[string]any
		if err := msgpack.Unmarshal(requestBody, &request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if request["name"] != "Onyx" || request["count"] != int64(7) {
			t.Errorf("request = %#v", request)
		}

		responseBody, err := msgpack.Marshal(map[string]any{"id": int64(91), "saved": true})
		if err != nil {
			t.Errorf("encode response: %v", err)
			return
		}
		w.Header().Set("Content-Type", msgpack.MediaType+"; version=1")
		_, _ = w.Write(responseBody)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	var response map[string]any
	err := c.DoEntity(
		context.Background(),
		http.MethodPut,
		"/data/db/widgets",
		map[string]any{"name": "Onyx", "count": int64(7)},
		&response,
		contract.WireFormatMessagePack,
	)
	if err != nil {
		t.Fatalf("DoEntity: %v", err)
	}
	if response["id"] != int64(91) || response["saved"] != true {
		t.Fatalf("response = %#v", response)
	}
}

func TestDoEntityMessagePackFallsBackToJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"records":[{"id":4}]}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	var response contract.QueryResults
	if err := c.DoEntity(context.Background(), http.MethodPut, "/query", map[string]any{}, &response, contract.WireFormatMessagePack); err != nil {
		t.Fatalf("DoEntity: %v", err)
	}
	if len(response) != 1 || response[0]["id"] != float64(4) {
		t.Fatalf("response = %#v", response)
	}
}

func TestDoEntityMessagePackBodylessRequestOmitsContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("bodyless Content-Type = %q, want empty", got)
		}
		if got := r.Header.Get("Accept"); got != msgpack.MediaType+", application/json;q=0.9" {
			t.Errorf("Accept = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	if err := c.DoEntity(
		context.Background(),
		http.MethodDelete,
		"/data/db/widgets/91",
		nil,
		nil,
		contract.WireFormatMessagePack,
	); err != nil {
		t.Fatalf("DoEntity: %v", err)
	}
}

func TestDoEntityMessagePackPreservesJSONErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"code":"invalid_entity","message":"invalid value","meta":{"field":"name"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	err := c.DoEntity(context.Background(), http.MethodPut, "/data/db/widgets", map[string]any{}, nil, contract.WireFormatMessagePack)
	var contractErr *contract.Error
	if !errors.As(err, &contractErr) {
		t.Fatalf("error = %T %v", err, err)
	}
	if contractErr.Code != "invalid_entity" || contractErr.Meta["status"] != http.StatusUnprocessableEntity || contractErr.Meta["field"] != "name" {
		t.Fatalf("contract error = %#v", contractErr)
	}
}

func TestDoEntityMessagePackParsesCloudJSONErrorEnvelope(t *testing.T) {
	const message = "Request body could not serialize a valid entity for entity widgets"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":{"message":"` + message + `"}}`))
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	err := c.DoEntity(context.Background(), http.MethodPut, "/data/db/widgets", map[string]any{}, nil, contract.WireFormatMessagePack)
	var contractErr *contract.Error
	if !errors.As(err, &contractErr) {
		t.Fatalf("error = %T %v", err, err)
	}
	if contractErr.Message != message || contractErr.Code != "" || contractErr.Meta["status"] != http.StatusUnprocessableEntity {
		t.Fatalf("contract error = %#v", contractErr)
	}
	if _, retainedRawBody := contractErr.Meta["body"]; retainedRawBody {
		t.Fatalf("recognized cloud error should not fall back to raw response body: %#v", contractErr)
	}
}

func TestDoEntityMessagePackLoggingDoesNotRenderBinary(t *testing.T) {
	var logs bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, _ := msgpack.Marshal(map[string]any{"secret-value": "must-not-be-logged"})
		w.Header().Set("Content-Type", msgpack.MediaType)
		_, _ = w.Write(response)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{
		Logger:       log.New(&logs, "", 0),
		LogRequests:  true,
		LogResponses: true,
	})
	var response map[string]any
	if err := c.DoEntity(context.Background(), http.MethodPut, "/data", map[string]any{"request-secret": "hidden"}, &response, contract.WireFormatMessagePack); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(logs.String(), "MessagePack body") {
		t.Fatalf("logs do not describe binary body: %s", logs.String())
	}
	if strings.Contains(logs.String(), "request-secret") || strings.Contains(logs.String(), "secret-value") {
		t.Fatalf("logs rendered binary payload: %s", logs.String())
	}
}

func TestDoEntityStreamMessagePackNegotiation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != msgpack.MediaType {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Accept") != msgpack.MediaType+", application/json;q=0.9" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", msgpack.MediaType)
		_, _ = w.Write([]byte{0xc0})
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client(), Options{})
	resp, err := c.DoEntityStream(context.Background(), http.MethodPut, "/stream", map[string]any{}, contract.WireFormatMessagePack)
	if err != nil {
		t.Fatalf("DoEntityStream: %v", err)
	}
	defer resp.Body.Close()
	if !IsMessagePackResponse(resp) {
		t.Fatal("expected MessagePack response")
	}
}
