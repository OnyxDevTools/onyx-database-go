package impl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestChatUsesDatabaseIDAndDisablesStream(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()

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

	body := strings.Join([]string{
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

func TestRequestScriptApprovalUsesDatabaseIDOverride(t *testing.T) {
	clearHTTPClientCache()
	ClearConfigCache()

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

func strPtr(s string) *string {
	return &s
}
