package impl

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func (q *query) queryPath() string {
	return "/data/" + url.PathEscape(q.client.cfg.DatabaseID) + "/query/" + url.PathEscape(q.table)
}

func (q *query) List(ctx context.Context) (contract.QueryResults, error) {
	payload := buildQueryPayload(q, true)
	var resp contract.QueryResults
	if err := q.client.httpClient.DoEntity(ctx, http.MethodPut, q.queryPath(), payload, &resp, q.client.wireFormat); err != nil {
		return nil, err
	}
	return resp, nil
}

func (q *query) Page(ctx context.Context, cursor string) (contract.PageResult, error) {
	payload := buildQueryPayload(q, false)
	params := url.Values{}
	if q.limit != nil && *q.limit > 0 {
		params.Set("pageSize", strconv.Itoa(*q.limit))
	}
	if cursor != "" {
		params.Set("nextPage", cursor)
	}

	var resp contract.PageResult
	path := q.queryPath()
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	if err := q.client.httpClient.DoEntity(ctx, http.MethodPut, path, payload, &resp, q.client.wireFormat); err != nil {
		return contract.PageResult{}, err
	}
	return resp, nil
}

func (q *query) Stream(ctx context.Context) (contract.Iterator, error) {
	payload := buildQueryPayload(q, true)
	path := "/data/" + url.PathEscape(q.client.cfg.DatabaseID) + "/query/stream/" + url.PathEscape(q.table)
	resp, err := q.client.httpClient.DoEntityStream(ctx, http.MethodPut, path, payload, q.client.wireFormat)
	if err != nil {
		return nil, err
	}
	return newStreamIterator(resp), nil
}

func (q *query) Update(ctx context.Context) (int, error) {
	payload := buildUpdatePayload(q)
	path := "/data/" + url.PathEscape(q.client.cfg.DatabaseID) + "/query/update/" + url.PathEscape(q.table)
	var updated int
	if err := q.client.httpClient.DoEntity(ctx, http.MethodPut, path, payload, &updated, q.client.wireFormat); err != nil {
		return 0, err
	}
	return updated, nil
}

func (q *query) Delete(ctx context.Context) (int, error) {
	payload := buildQueryPayload(q, true)
	path := "/data/" + url.PathEscape(q.client.cfg.DatabaseID) + "/query/delete/" + url.PathEscape(q.table)
	var deleted int
	if err := q.client.httpClient.DoEntity(ctx, http.MethodPut, path, payload, &deleted, q.client.wireFormat); err != nil {
		return 0, err
	}
	return deleted, nil
}
