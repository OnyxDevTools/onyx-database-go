package impl

import (
	"context"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type cascadeClient struct {
	client *client
	spec   contract.CascadeSpec
}

func (c *cascadeClient) Save(ctx context.Context, table string, entity any) error {
	_, err := c.client.Save(ctx, table, entity, []string{c.spec.String()})
	return err
}

func (c *cascadeClient) Delete(ctx context.Context, table, id string) error {
	// Cascade deletes reuse the same Delete endpoint with relationship spec.
	// The API currently only supports relationship filters via resolvers; no extra payload needed.
	return c.client.Delete(ctx, table, id)
}
