package contract

import "context"

// Secret represents a key-value secret stored in Onyx.
type Secret struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// SecretClient exposes the Secrets API surface.
type SecretClient interface {
	List(ctx context.Context) ([]Secret, error)
	Get(ctx context.Context, key string) (Secret, error)
	Set(ctx context.Context, secret Secret) (Secret, error)
	Delete(ctx context.Context, key string) error
}
