package onyx

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func buildMetricsPath(params contract.MetricsParams) (string, error) {
	if params.UID == "" && params.Email == "" {
		return "", errors.New("metrics requires a uid or email")
	}

	values := url.Values{}
	if params.UID != "" {
		values.Set("uid", params.UID)
	}
	if params.Email != "" {
		values.Set("email", params.Email)
	}

	return "/metrics?" + values.Encode(), nil
}

func (c *client) Metrics(ctx context.Context, params contract.MetricsParams) (contract.Metrics, error) {
	path, err := buildMetricsPath(params)
	if err != nil {
		return nil, err
	}

	var metrics contract.Metrics
	if err := c.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}
