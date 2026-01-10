package contract

import "context"

// OnyxDocument represents a stored document payload.
// The structure is intentionally flexible and mirrors the TS client shape.
type OnyxDocument struct {
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

// OnyxDocumentsClient exposes the Documents API surface.
type OnyxDocumentsClient interface {
	List(ctx context.Context) ([]OnyxDocument, error)
	Get(ctx context.Context, id string) (OnyxDocument, error)
	Save(ctx context.Context, doc OnyxDocument) (OnyxDocument, error)
	Delete(ctx context.Context, id string) error
}

// Backward-compatible aliases to avoid breaking existing callers.
type Document = OnyxDocument
type DocumentClient = OnyxDocumentsClient
