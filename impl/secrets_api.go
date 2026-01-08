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

func (c *client) Secrets() contract.SecretClient { return &secretClient{client: c} }

type secretListResponse struct {
	Records []contract.Secret `json:"records"`
}

func (s *secretClient) List(ctx context.Context) ([]contract.Secret, error) {
	var resp secretListResponse
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret"
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Records, nil
}

func (s *secretClient) Get(ctx context.Context, key string) (contract.Secret, error) {
	if key == "" {
		return contract.Secret{}, fmt.Errorf("secret key is required")
	}
	var secret contract.Secret
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret/" + tableEscape(key)
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &secret); err != nil {
		return contract.Secret{}, err
	}
	return secret, nil
}

func (s *secretClient) Set(ctx context.Context, secret contract.Secret) (contract.Secret, error) {
	if secret.Key == "" {
		return contract.Secret{}, fmt.Errorf("secret key is required")
	}
	path := "/database/" + tableEscape(s.client.cfg.DatabaseID) + "/secret/" + tableEscape(secret.Key)
	var resp contract.Secret
	if err := s.client.httpClient.DoJSON(ctx, http.MethodPut, path, secret, &resp); err != nil {
		return contract.Secret{}, err
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
func (c *client) ListSecrets(ctx context.Context) ([]contract.Secret, error) {
	return c.Secrets().List(ctx)
}

func (c *client) GetSecret(ctx context.Context, key string) (contract.Secret, error) {
	return c.Secrets().Get(ctx, key)
}

func (c *client) PutSecret(ctx context.Context, secret contract.Secret) (contract.Secret, error) {
	return c.Secrets().Set(ctx, secret)
}

func (c *client) DeleteSecret(ctx context.Context, key string) error {
	return c.Secrets().Delete(ctx, key)
}
