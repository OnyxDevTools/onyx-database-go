package impl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type secretClient struct {
	client *client
}

func (c *client) OnyxSecrets() contract.OnyxSecretsClient { return &secretClient{client: c} }
func (c *client) OnyxSecret() contract.OnyxSecretsClient  { return c.OnyxSecrets() }
func (c *client) Secrets() contract.OnyxSecretsClient     { return c.OnyxSecrets() }

type secretListResponse struct {
	Records []contract.OnyxSecret `json:"records"`
}

func (s *secretClient) List(ctx context.Context) ([]contract.OnyxSecret, error) {
	var resp secretListResponse
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret"
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Records, nil
}

func (s *secretClient) Get(ctx context.Context, key string) (contract.OnyxSecret, error) {
	if key == "" {
		return contract.OnyxSecret{}, fmt.Errorf("secret key is required")
	}
	var secret contract.OnyxSecret
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret/" + tableEscape(key)
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &secret); err != nil {
		return contract.OnyxSecret{}, err
	}
	return secret, nil
}

func (s *secretClient) Set(ctx context.Context, secret contract.OnyxSecret) (contract.OnyxSecret, error) {
	if secret.Key == "" {
		return contract.OnyxSecret{}, fmt.Errorf("secret key is required")
	}
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret/" + tableEscape(secret.Key)
	var resp contract.OnyxSecret
	if err := s.client.httpClient.DoJSON(ctx, http.MethodPut, path, secret, &resp); err != nil {
		return contract.OnyxSecret{}, err
	}
	return resp, nil
}

func (s *secretClient) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("secret key is required")
	}
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret/" + tableEscape(key)
	return s.client.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}

// Facade-style helpers matching the TS SDK surface.
func (c *client) ListSecrets(ctx context.Context) ([]contract.OnyxSecret, error) {
	return c.Secrets().List(ctx)
}

func (c *client) GetSecret(ctx context.Context, key string) (contract.OnyxSecret, error) {
	return c.Secrets().Get(ctx, key)
}

func (c *client) PutSecret(ctx context.Context, secret contract.OnyxSecret) (contract.OnyxSecret, error) {
	return c.Secrets().Set(ctx, secret)
}

func (c *client) DeleteSecret(ctx context.Context, key string) error {
	return c.Secrets().Delete(ctx, key)
}
