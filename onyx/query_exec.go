package onyx

import (
	"context"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func (q *query) List(ctx context.Context) (contract.QueryResults, error) {
	payload := buildQueryPayload(q)
	var resp contract.QueryResults
	if err := q.client.httpClient.DoJSON(ctx, http.MethodPost, "/query/list", payload, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (q *query) Page(ctx context.Context, cursor string) (contract.PageResult, error) {
	payload := struct {
		queryPayload
		Cursor string `json:"cursor,omitempty"`
	}{queryPayload: buildQueryPayload(q), Cursor: cursor}

	var resp contract.PageResult
	if err := q.client.httpClient.DoJSON(ctx, http.MethodPost, "/query/page", payload, &resp); err != nil {
		return contract.PageResult{}, err
	}
	return resp, nil
}

func (q *query) Stream(ctx context.Context) (contract.Iterator, error) {
	payload := buildQueryPayload(q)
	resp, err := q.client.httpClient.DoStream(ctx, http.MethodPost, "/query/stream", payload)
	if err != nil {
		return nil, err
	}
	return newStreamIterator(resp), nil
}
