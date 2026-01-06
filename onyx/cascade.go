package onyx

import (
	"context"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type cascadeClient struct {
	client *client
	spec   contract.CascadeSpec
}

func (c *cascadeClient) Save(ctx context.Context, table string, entity any) error {
	payload := map[string]any{
		"spec":   c.spec.String(),
		"entity": entity,
	}
	path := "/cascade/" + table + "/save"
	return c.client.httpClient.DoJSON(ctx, http.MethodPost, path, payload, nil)
}

func (c *cascadeClient) Delete(ctx context.Context, table, id string) error {
	payload := map[string]any{
		"spec": c.spec.String(),
	}
	path := "/cascade/" + table + "/" + id
	return c.client.httpClient.DoJSON(ctx, http.MethodDelete, path, payload, nil)
}
