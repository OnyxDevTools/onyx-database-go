package onyx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type secretClient struct {
	client *client
}

func (c *client) Secrets() contract.SecretClient {
	return &secretClient{client: c}
}

func (s *secretClient) List(ctx context.Context) ([]contract.Secret, error) {
	var secrets []contract.Secret
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, "/secrets", nil, &secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *secretClient) Get(ctx context.Context, key string) (contract.Secret, error) {
	if key == "" {
		return contract.Secret{}, fmt.Errorf("secret key is required")
	}
	var secret contract.Secret
	path := "/secrets/" + key
	if err := s.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &secret); err != nil {
		return contract.Secret{}, err
	}
	return secret, nil
}

func (s *secretClient) Set(ctx context.Context, secret contract.Secret) (contract.Secret, error) {
	if secret.Key == "" {
		return contract.Secret{}, fmt.Errorf("secret key is required")
	}
	path := "/secrets/" + secret.Key
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
	path := "/secrets/" + key
	return s.client.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}
