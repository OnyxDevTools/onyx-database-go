package contract

import "context"

// Document represents a stored document payload.
// The structure is intentionally flexible and mirrors the TS client shape.
type Document struct {
	// Legacy ID used by earlier iterations of the Go SDK.
	ID string `json:"id,omitempty"`
	// Canonical identifier used by the TS client and HTTP API.
	DocumentID string `json:"documentId,omitempty"`
	// Optional path within the database.
	Path string `json:"path,omitempty"`
	// MIME type for the stored content.
	MimeType string `json:"mimeType,omitempty"`
	// Base64-encoded document content.
	Content string `json:"content,omitempty"`
	// Generic key/value payload preserved for backwards compatibility.
	Data map[string]any `json:"data,omitempty"`
	// Optional timestamps the service may return.
	Created   string `json:"created,omitempty"`
	Updated   string `json:"updated,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// DocumentClient exposes the Documents API surface.
type DocumentClient interface {
	List(ctx context.Context) ([]Document, error)
	Get(ctx context.Context, id string) (Document, error)
	Save(ctx context.Context, doc Document) (Document, error)
	Delete(ctx context.Context, id string) error
}
