package impl

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
	"github.com/OnyxDevTools/onyx-database-go/internal/httpclient"
)

func TestBatchSaveDefaultSizeAndRetryFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, `{"code":"rate","message":"slow"}`, http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
	}

	err := batchSave(context.Background(), c, "users", []any{1, 2, 3}, 0)
	if err == nil {
		t.Fatalf("expected retry failure")
	}
	if calls != 2 {
		t.Fatalf("expected two attempts with retry, got %d", calls)
	}
}

func TestBatchSaveUsesDefaultBatchSizeNegative(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
	}

	if err := batchSave(context.Background(), c, "users", []any{"a", "b"}, -1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected single call with default batch size, got %d", calls)
	}
}

func TestBatchSaveCancelDuringRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	origHook := testHookBeforeRetryWait
	testHookBeforeRetryWait = func(context.Context) {
		cancel()
	}
	defer func() { testHookBeforeRetryWait = origHook }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"rate","message":"slow"}`, http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	c := &client{
		httpClient: httpclient.New(srv.URL, srv.Client(), httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
	}
	if err := batchSave(ctx, c, "users", []any{1}, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancel, got %v", err)
	}
}

func TestGetCachedHTTPClientLoadedPath(t *testing.T) {
	clearHTTPClientCache()
	signer := httpclient.Signer{APIKey: "k", APISecret: "s"}
	logger := log.New(io.Discard, "", 0)
	baseHTTP := &http.Client{}

	start := make(chan struct{})
	var c1, c2 *httpclient.Client
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		c1 = getCachedHTTPClient("http://example.com", baseHTTP, false, false, signer, logger)
	}()
	go func() {
		defer wg.Done()
		<-start
		c2 = getCachedHTTPClient("http://example.com", baseHTTP, false, false, signer, logger)
	}()
	close(start)
	wg.Wait()

	if c1 == nil || c2 == nil || c1 != c2 {
		t.Fatalf("expected shared cached client")
	}
}

func TestGetCachedHTTPClientLoadedPathViaHook(t *testing.T) {
	clearHTTPClientCache()
	defer func() { testHookAfterCacheLoad = nil }()

	ready := make(chan struct{}, 2)
	release := make(chan struct{})
	testHookAfterCacheLoad = func() {
		ready <- struct{}{}
		<-release
	}

	signer := httpclient.Signer{APIKey: "k", APISecret: "s"}
	logger := log.New(io.Discard, "", 0)
	baseHTTP := &http.Client{}

	var c1, c2 *httpclient.Client
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		c1 = getCachedHTTPClient("http://example.com", baseHTTP, false, false, signer, logger)
	}()
	go func() {
		defer wg.Done()
		c2 = getCachedHTTPClient("http://example.com", baseHTTP, false, false, signer, logger)
	}()

	<-ready
	<-ready
	close(release)
	wg.Wait()

	if c1 == nil || c2 == nil || c1 != c2 {
		t.Fatalf("expected shared cached client after load-or-store")
	}
}

func TestInitRespectsCustomClockAndSleep(t *testing.T) {
	ClearConfigCache()
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	slept := false
	cfg := Config{
		DatabaseID:      "db_clock",
		DatabaseBaseURL: "http://example.com",
		APIKey:          "k",
		APISecret:       "s",
		Clock: func() time.Time {
			return now
		},
		Sleep: func(time.Duration) {
			slept = true
		},
	}
	c, err := Init(context.Background(), cfg)
	if err != nil {
		t.Fatalf("init err: %v", err)
	}
	implClient := c.(*client)
	if implClient.now().UTC() != now {
		t.Fatalf("expected custom clock to be used")
	}
	implClient.sleep(time.Millisecond)
	if !slept {
		t.Fatalf("expected custom sleep to be used")
	}
}

func TestInitErrorOnMissingConfig(t *testing.T) {
	ClearConfigCache()
	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")

	if _, err := Init(context.Background(), Config{}); err == nil {
		t.Fatalf("expected init to fail when config missing")
	}
}

func TestClientSaveErrorPropagation(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})
	if _, err := c.Save(context.Background(), "users", map[string]any{"id": 1}, nil); err == nil {
		t.Fatalf("expected save error to propagate")
	}
}

func TestDocumentSaveFillsMissingIDs(t *testing.T) {
	var body map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["id"] != "doc123" || body["documentId"] != "doc123" {
			t.Fatalf("expected ids set from preferred document id, got %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"documentId":"doc123","id":""}`))
	})

	doc, err := c.Documents().Save(context.Background(), contract.OnyxDocument{DocumentID: "doc123"})
	if err != nil {
		t.Fatalf("save err: %v", err)
	}
	if doc.ID != "doc123" || doc.DocumentID != "doc123" {
		t.Fatalf("expected ids normalized: %+v", doc)
	}
}

func TestDocumentSaveError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusBadRequest)
	})
	if _, err := c.Documents().Save(context.Background(), contract.OnyxDocument{DocumentID: "doc123"}); err == nil {
		t.Fatalf("expected document save to propagate error")
	}
}

func TestBuildUpdatePayloadWithLimit(t *testing.T) {
	limit := 2
	q := &query{table: "users", updates: map[string]any{"x": 1}, limit: &limit}
	payload := buildUpdatePayload(q)
	if payload.Limit == nil || *payload.Limit != 2 {
		t.Fatalf("expected limit set in update payload")
	}
}

func TestStreamIteratorSkipsBlankLines(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(bufio.NewReader(strings.NewReader("\n{\"id\":1}\n"))),
	}
	iter := newStreamIterator(resp)
	if !iter.Next() {
		t.Fatalf("expected next to skip blank and return data")
	}
	val := iter.Value()
	if val["id"] != float64(1) || iter.Err() != nil {
		t.Fatalf("unexpected value or error: %+v err=%v", val, iter.Err())
	}
	iter.Close()
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestStreamIteratorScannerError(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(errorReader{})}
	iter := newStreamIterator(resp)
	if iter.Next() {
		t.Fatalf("expected Next to fail with scanner error")
	}
	if !errors.Is(iter.Err(), io.ErrUnexpectedEOF) {
		t.Fatalf("expected scanner error, got %v", iter.Err())
	}
	iter.Close()
}

func TestFetchSchemaErrorNotFoundOnlyTriggersFallback(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusInternalServerError)
	})
	if _, err := fetchSchema(context.Background(), c, nil); err == nil {
		t.Fatalf("expected fetchSchema to return non-404 error")
	}
}

func TestFetchSchemaFallbackSchemasSuccess(t *testing.T) {
	call := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		call++
		switch call {
		case 1:
			http.Error(w, `{"code":"missing","message":"no schema"}`, http.StatusNotFound)
		case 2:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"entities":[{"name":"A","attributes":[{"name":"id","type":"String"}]}]}`))
		default:
			t.Fatalf("unexpected call %d", call)
		}
	})

	schema, err := fetchSchema(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchSchema err: %v", err)
	}
	if _, ok := schema.Table("A"); !ok {
		t.Fatalf("expected table parsed from entities fallback")
	}
}

func TestFetchSchemaSchemaObjectEntities(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"schema":{"entities":[{"name":"B","attributes":[{"name":"id","type":"String"}]}]}}`))
	})
	schema, err := fetchSchema(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchSchema err: %v", err)
	}
	if _, ok := schema.Table("B"); !ok {
		t.Fatalf("expected schema object entities parsed")
	}
}

type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("boom")
}

func TestFetchSchemaNonContractError(t *testing.T) {
	hc := &http.Client{Transport: errRoundTripper{}}
	c := &client{
		httpClient: httpclient.New("http://example.com", hc, httpclient.Options{}),
		cfg:        resolver.ResolvedConfig{DatabaseID: "db"},
	}
	if _, err := fetchSchema(context.Background(), c, nil); err == nil {
		t.Fatalf("expected non-contract error to propagate")
	}
}

func TestFetchSchemaFallbackLegacyError(t *testing.T) {
	call := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		call++
		switch call {
		case 1:
			http.Error(w, `{"code":"missing","message":"no schema"}`, http.StatusNotFound)
		case 2:
			http.Error(w, `{"code":"bad","message":"fail"}`, http.StatusInternalServerError)
		default:
			http.Error(w, `{"code":"legacy","message":"fail"}`, http.StatusInternalServerError)
		}
	})
	if _, err := fetchSchema(context.Background(), c, nil); err == nil {
		t.Fatalf("expected error when all fallbacks fail")
	}
	if call != 3 {
		t.Fatalf("expected three calls, got %d", call)
	}
}

func TestFetchSchemaMarshalError(t *testing.T) {
	origMarshal := jsonMarshal
	defer func() { jsonMarshal = origMarshal }()
	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal boom")
	}

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"unknown":1}`))
	})
	if _, err := fetchSchema(context.Background(), c, nil); err == nil {
		t.Fatalf("expected marshal error to propagate")
	}
}

func TestSchemaFromEntitiesAndTablesSkipInvalid(t *testing.T) {
	entities := []any{
		123,
		map[string]any{
			"name":       "Thing",
			"identifier": map[string]any{"name": "id"},
			"attributes": []any{"bad-attr"},
			"resolvers":  []any{123},
		},
	}
	schema := schemaFromEntities(entities)
	if len(schema.Tables) != 1 || len(schema.Tables[0].Fields) != 0 || len(schema.Tables[0].Resolvers) != 0 {
		t.Fatalf("expected invalid entries skipped, got %+v", schema.Tables)
	}

	tables := []any{
		456,
		map[string]any{
			"name":      "Docs",
			"fields":    []any{123},
			"resolvers": []any{123},
		},
	}
	schema2 := schemaFromTablesArray(tables)
	if len(schema2.Tables) != 1 || len(schema2.Tables[0].Fields) != 0 || len(schema2.Tables[0].Resolvers) != 0 {
		t.Fatalf("expected invalid table entries skipped, got %+v", schema2.Tables)
	}
}
