package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type errReader struct{ called bool }

func (r *errReader) Read(p []byte) (int, error) {
	if !r.called {
		r.called = true
		data := []byte("data: {\"id\":\"chunk-1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hi\"}}]}\n")
		copy(p, data)
		return len(data), nil
	}
	return 0, fmt.Errorf("boom")
}

func (r *errReader) Close() error { return nil }

func TestChatUsesDatabaseIDAndDisablesStream(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	var capturedQuery string
	var capturedStream any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		capturedQuery = r.URL.RawQuery
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		capturedStream = payload["stream"]
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contract.AIChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1,
			Model:   "onyx-chat",
			Choices: []contract.AIChatCompletionChoice{
				{Index: 0, Message: contract.AIChatMessage{Role: "assistant", Content: "hi"}, FinishReason: strPtr("stop")},
			},
		})
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "db123",
		DatabaseBaseURL: server.URL,
		AIBaseURL:       server.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	resp, err := client.Chat(ctx, contract.AIChatCompletionRequest{
		Model:    "onyx-chat",
		Messages: []contract.AIChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat call failed: %v", err)
	}
	if resp.ID != "chatcmpl-123" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !strings.Contains(capturedQuery, "databaseId=db123") {
		t.Fatalf("databaseId not forwarded in query: %s", capturedQuery)
	}
	if capturedStream != nil {
		t.Fatalf("expected stream to be omitted/false, got %v", capturedStream)
	}
}

func TestChatStreamSetsStreamAndParsesChunks(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	body := strings.Join([]string{
		"",
		"event: ping",
		`data: {"id":"chunk-1","object":"chat.completion.chunk","created":1,"model":"onyx-chat","choices":[{"index":0,"delta":{"content":"hi"},"finish_reason":null}]}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["stream"] != true {
			t.Fatalf("expected stream=true, got %v", payload["stream"])
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "db123",
		DatabaseBaseURL: server.URL,
		AIBaseURL:       server.URL,
		APIKey:          "k",
		APISecret:       "s",
		LogResponses:    true,
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	stream, err := client.ChatStream(ctx, contract.AIChatCompletionRequest{
		Model:    "onyx-chat",
		Messages: []contract.AIChatMessage{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat stream failed: %v", err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("expected first chunk")
	}
	chunk := stream.Chunk()
	if chunk.ID != "chunk-1" || chunk.Model != "onyx-chat" {
		t.Fatalf("unexpected chunk: %+v", chunk)
	}
	if stream.Next() {
		t.Fatalf("expected stream to end after [DONE]")
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("unexpected stream error: %v", err)
	}
}

func TestChatStreamHandlesScannerErrorAndLogs(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	resp := &http.Response{StatusCode: 200, Body: &errReader{}}
	loggerBuf := &bytes.Buffer{}
	stream := newAIChatStream(resp, log.New(loggerBuf, "", 0), true)

	if !stream.Next() {
		t.Fatalf("expected first chunk")
	}
	if stream.Chunk().ID != "chunk-1" {
		t.Fatalf("chunk not parsed")
	}
	if stream.Next() {
		t.Fatalf("expected error on second Next")
	}
	if stream.Next() {
		t.Fatalf("expected early stop when err is set")
	}
	if stream.Err() == nil {
		t.Fatalf("expected stored error")
	}
	if loggerBuf.Len() == 0 {
		t.Fatalf("expected logging when logResponses enabled")
	}
}

func TestChatStreamInvalidJSONSetsError(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("data: not-json\n")),
	}
	stream := newAIChatStream(resp, nil, false)
	if stream.Next() {
		t.Fatalf("expected Next to fail on invalid JSON")
	}
	if stream.Err() == nil {
		t.Fatalf("expected unmarshal error")
	}
}

func TestRequestScriptApprovalUsesDatabaseIDOverride(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/script-approvals" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(contract.AIScriptApprovalResponse{
			NormalizedScript: "norm",
			ExpiresAtIso:     "2024-01-01T00:00:00Z",
			RequiresApproval: true,
			Findings:         "ok",
		})
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "cfg-db",
		DatabaseBaseURL: server.URL,
		AIBaseURL:       server.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	resp, err := client.RequestScriptApproval(ctx, contract.AIScriptApprovalRequest{
		Script:     "save()",
		DatabaseID: "req-db",
	})
	if err != nil {
		t.Fatalf("script approval failed: %v", err)
	}
	if resp.NormalizedScript != "norm" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !strings.Contains(capturedQuery, "databaseId=req-db") {
		t.Fatalf("expected request databaseId to override config, got %s", capturedQuery)
	}
}

func TestRequestScriptApprovalFallsBackToConfigAndErrors(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "cfg-db",
		DatabaseBaseURL: server.URL,
		AIBaseURL:       server.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	_, err = client.RequestScriptApproval(ctx, contract.AIScriptApprovalRequest{
		Script: "save()",
		// DatabaseID intentionally empty to force fallback
	})
	if err == nil {
		t.Fatalf("expected error from DoJSON")
	}
	if !strings.Contains(capturedQuery, "databaseId=cfg-db") {
		t.Fatalf("expected config databaseId in query, got %s", capturedQuery)
	}
}

func TestGetModelsAndGetModelSuccessAndError(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	success := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			_ = json.NewEncoder(w).Encode(contract.AIModelsResponse{Data: []contract.AIModel{{ID: "m1"}}})
		case "/v1/models/m2":
			_ = json.NewEncoder(w).Encode(contract.AIModel{ID: "m2"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer success.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "db123",
		DatabaseBaseURL: success.URL,
		AIBaseURL:       success.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	models, err := client.GetModels(ctx)
	if err != nil || len(models.Data) != 1 || models.Data[0].ID != "m1" {
		t.Fatalf("unexpected models response %+v, err=%v", models, err)
	}
	model, err := client.GetModel(ctx, "m2")
	if err != nil || model.ID != "m2" {
		t.Fatalf("unexpected model response %+v, err=%v", model, err)
	}

	errorSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer errorSrv.Close()

	errClient, err := Init(ctx, Config{
		DatabaseID:      "db123",
		DatabaseBaseURL: errorSrv.URL,
		AIBaseURL:       errorSrv.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init err client: %v", err)
	}

	if _, err := errClient.GetModels(ctx); err == nil {
		t.Fatalf("expected GetModels error")
	}
	if _, err := errClient.GetModel(ctx, "m2"); err == nil {
		t.Fatalf("expected GetModel error")
	}
}

func TestChatErrorAndNoDatabaseID(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer errSrv.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "db",
		DatabaseBaseURL: errSrv.URL,
		AIBaseURL:       errSrv.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	_, err = client.Chat(ctx, contract.AIChatCompletionRequest{
		Model:    "onyx-chat",
		Messages: []contract.AIChatMessage{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatalf("expected chat error")
	}
}

func TestChatStreamErrorPath(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()
	clearEnv(t)

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer errSrv.Close()

	ctx := context.Background()
	client, err := Init(ctx, Config{
		DatabaseID:      "db",
		DatabaseBaseURL: errSrv.URL,
		AIBaseURL:       errSrv.URL,
		APIKey:          "k",
		APISecret:       "s",
	})
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	if _, err := client.ChatStream(ctx, contract.AIChatCompletionRequest{
		Model:    "onyx-chat",
		Messages: []contract.AIChatMessage{{Role: "user", Content: "hi"}},
	}); err == nil {
		t.Fatalf("expected ChatStream error")
	}
}

func TestChooseDatabaseID(t *testing.T) {
	tests := []struct {
		name      string
		requested string
		fallback  string
		want      string
	}{
		{"requested wins", "  req  ", "fb", "req"},
		{"fallback", "", " fb ", "fb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chooseDatabaseID(tt.requested, tt.fallback); got != tt.want {
				t.Fatalf("chooseDatabaseID = %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func clearEnv(t *testing.T) {
	t.Helper()
	t.Setenv("ONYX_DATABASE_ID", "")
	t.Setenv("ONYX_DATABASE_BASE_URL", "")
	t.Setenv("ONYX_DATABASE_API_KEY", "")
	t.Setenv("ONYX_DATABASE_API_SECRET", "")
	t.Setenv("ONYX_CONFIG_PATH", "")
	t.Setenv("ONYX_AI_BASE_URL", "")
}
