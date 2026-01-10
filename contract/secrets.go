package contract

import "context"

// OnyxSecret represents a key-value secret stored in Onyx.
type OnyxSecret struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// OnyxSecretsClient exposes the Secrets API surface.
type OnyxSecretsClient interface {
	List(ctx context.Context) ([]OnyxSecret, error)
	Get(ctx context.Context, key string) (OnyxSecret, error)
	Set(ctx context.Context, secret OnyxSecret) (OnyxSecret, error)
	Delete(ctx context.Context, key string) error
}

// Backward-compatible aliases to avoid breaking existing callers.
type Secret = OnyxSecret
type SecretClient = OnyxSecretsClient
