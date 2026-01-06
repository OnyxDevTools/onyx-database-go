package contract

import "context"

// Document represents a stored document payload.
// The structure is intentionally flexible and mirrors the TS client shape.
type Document struct {
	ID        string         `json:"id"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"createdAt,omitempty"`
	UpdatedAt string         `json:"updatedAt,omitempty"`
}

// DocumentClient exposes the Documents API surface.
type DocumentClient interface {
	List(ctx context.Context) ([]Document, error)
	Get(ctx context.Context, id string) (Document, error)
	Save(ctx context.Context, doc Document) (Document, error)
	Delete(ctx context.Context, id string) error
}
