package impl

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func (c *client) Chat(ctx context.Context, req contract.AIChatCompletionRequest) (contract.AIChatCompletionResponse, error) {
	req.Stream = false
	path := "/v1/chat/completions"
	if dbID := chooseDatabaseID(req.DatabaseID, c.cfg.DatabaseID); dbID != "" {
		params := url.Values{}
		params.Set("databaseId", dbID)
		path += "?" + params.Encode()
	}

	var resp contract.AIChatCompletionResponse
	if err := c.aiClient.DoJSON(ctx, http.MethodPost, path, req, &resp); err != nil {
		return contract.AIChatCompletionResponse{}, err
	}
	return resp, nil
}

func (c *client) ChatStream(ctx context.Context, req contract.AIChatCompletionRequest) (contract.AIChatStream, error) {
	req.Stream = true
	path := "/v1/chat/completions"
	if dbID := chooseDatabaseID(req.DatabaseID, c.cfg.DatabaseID); dbID != "" {
		params := url.Values{}
		params.Set("databaseId", dbID)
		path += "?" + params.Encode()
	}

	resp, err := c.aiClient.DoStream(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	return newAIChatStream(resp, c.aiClient.Logger(), c.aiClient.LogResponses()), nil
}

func (c *client) GetModels(ctx context.Context) (contract.AIModelsResponse, error) {
	var resp contract.AIModelsResponse
	if err := c.aiClient.DoJSON(ctx, http.MethodGet, "/v1/models", nil, &resp); err != nil {
		return contract.AIModelsResponse{}, err
	}
	return resp, nil
}

func (c *client) GetModel(ctx context.Context, modelID string) (contract.AIModel, error) {
	path := "/v1/models/" + url.PathEscape(modelID)
	var resp contract.AIModel
	if err := c.aiClient.DoJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return contract.AIModel{}, err
	}
	return resp, nil
}

func (c *client) RequestScriptApproval(ctx context.Context, req contract.AIScriptApprovalRequest) (contract.AIScriptApprovalResponse, error) {
	path := "/api/script-approvals"
	if dbID := strings.TrimSpace(req.DatabaseID); dbID == "" {
		req.DatabaseID = c.cfg.DatabaseID
	}
	if strings.TrimSpace(req.DatabaseID) != "" {
		params := url.Values{}
		params.Set("databaseId", strings.TrimSpace(req.DatabaseID))
		path += "?" + params.Encode()
	}

	var resp contract.AIScriptApprovalResponse
	if err := c.aiClient.DoJSON(ctx, http.MethodPost, path, req, &resp); err != nil {
		return contract.AIScriptApprovalResponse{}, err
	}
	return resp, nil
}

func chooseDatabaseID(requested, fallback string) string {
	if trimmed := strings.TrimSpace(requested); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}
