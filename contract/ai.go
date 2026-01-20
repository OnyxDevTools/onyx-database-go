package contract

import "context"

// AIChatMessage represents a single message in a chat completion request or response.
type AIChatMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []AIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	Name       string       `json:"name,omitempty"`
}

// AITool describes a callable tool exposed to the model.
type AITool struct {
	Type     string         `json:"type"`
	Function AIToolFunction `json:"function"`
}

// AIToolFunction describes the function shape for a tool call.
type AIToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// AIToolCall represents a tool invocation returned by the model.
type AIToolCall struct {
	ID       string             `json:"id,omitempty"`
	Type     string             `json:"type,omitempty"`
	Function AIToolCallFunction `json:"function"`
}

// AIToolCallFunction describes the function call details within a tool call.
type AIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// AIChatCompletionRequest matches the OpenAI-style chat completion request.
type AIChatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []AIChatMessage `json:"messages"`
	Stream      bool            `json:"stream,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	Tools       []AITool        `json:"tools,omitempty"`
	ToolChoice  any             `json:"tool_choice,omitempty"`
	User        string          `json:"user,omitempty"`
	DatabaseID  string          `json:"-"`
}

// AIChatCompletionResponse represents a non-streaming chat completion response.
type AIChatCompletionResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []AIChatCompletionChoice `json:"choices"`
	Usage   *AIChatCompletionUsage   `json:"usage,omitempty"`
}

// AIChatCompletionChoice captures a single completion choice.
type AIChatCompletionChoice struct {
	Index        int           `json:"index"`
	Message      AIChatMessage `json:"message"`
	FinishReason *string       `json:"finish_reason,omitempty"`
}

// AIChatCompletionUsage reports token usage for the completion.
type AIChatCompletionUsage struct {
	PromptTokens     *int `json:"prompt_tokens,omitempty"`
	CompletionTokens *int `json:"completion_tokens,omitempty"`
	TotalTokens      *int `json:"total_tokens,omitempty"`
}

// AIChatCompletionChunk represents a streaming chat completion chunk.
type AIChatCompletionChunk struct {
	ID      string                        `json:"id"`
	Object  string                        `json:"object"`
	Created int64                         `json:"created"`
	Model   string                        `json:"model,omitempty"`
	Choices []AIChatCompletionChunkChoice `json:"choices"`
}

// AIChatCompletionChunkChoice captures a single streaming choice delta.
type AIChatCompletionChunkChoice struct {
	Index        int                        `json:"index"`
	Delta        AIChatCompletionChunkDelta `json:"delta"`
	FinishReason *string                    `json:"finish_reason,omitempty"`
}

// AIChatCompletionChunkDelta holds partial message updates in a stream.
type AIChatCompletionChunkDelta struct {
	Role       string       `json:"role,omitempty"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []AIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	Name       string       `json:"name,omitempty"`
}

// AIChatStream provides streaming access to chat completion chunks.
type AIChatStream interface {
	Next() bool
	Chunk() AIChatCompletionChunk
	Err() error
	Close() error
}

// AIModel describes a single AI model.
type AIModel struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// AIModelsResponse lists available models.
type AIModelsResponse struct {
	Object string    `json:"object"`
	Data   []AIModel `json:"data"`
}

// AIScriptApprovalRequest requests mutation approval for a script payload.
type AIScriptApprovalRequest struct {
	Script     string `json:"script"`
	DatabaseID string `json:"-"`
}

// AIScriptApprovalResponse reports whether a script requires approval.
type AIScriptApprovalResponse struct {
	NormalizedScript string `json:"normalizedScript"`
	ExpiresAtIso     string `json:"expiresAtIso"`
	RequiresApproval bool   `json:"requiresApproval"`
	Findings         string `json:"findings"`
}

// AIClient defines AI operations available from the SDK.
type AIClient interface {
	Chat(ctx context.Context, req AIChatCompletionRequest) (AIChatCompletionResponse, error)
	ChatStream(ctx context.Context, req AIChatCompletionRequest) (AIChatStream, error)
	GetModels(ctx context.Context) (AIModelsResponse, error)
	GetModel(ctx context.Context, modelID string) (AIModel, error)
	RequestScriptApproval(ctx context.Context, req AIScriptApprovalRequest) (AIScriptApprovalResponse, error)
}
