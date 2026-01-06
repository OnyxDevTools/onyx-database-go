package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta"`
}

func parseError(ctx context.Context, status int, body []byte) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}

	var payload errorPayload
	if err := json.Unmarshal(body, &payload); err == nil && (payload.Code != "" || payload.Message != "") {
		if payload.Meta == nil {
			payload.Meta = map[string]any{}
		}
		payload.Meta["status"] = status
		return &contract.Error{Code: payload.Code, Message: payload.Message, Meta: payload.Meta}
	}

	meta := map[string]any{"status": status}
	if len(body) > 0 {
		meta["body"] = string(body)
	}
	msg := http.StatusText(status)
	if msg == "" {
		msg = "request failed"
	}
	return &contract.Error{Message: msg, Meta: meta}
}
