package impl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type documentClient struct {
	client *client
}

func (c *client) Documents() contract.DocumentClient {
	return &documentClient{client: c}
}

func (d *documentClient) List(ctx context.Context) ([]contract.Document, error) {
	path := d.basePath()
	var docs []contract.Document
	if err := d.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &docs); err != nil {
		return nil, err
	}
	for i := range docs {
		docs[i] = normalizeDocumentIDs(docs[i])
	}
	return docs, nil
}

func (d *documentClient) Get(ctx context.Context, id string) (contract.Document, error) {
	docID := strings.TrimSpace(id)
	if docID == "" {
		return contract.Document{}, fmt.Errorf("document id is required")
	}

	var doc contract.Document
	path := d.basePath() + "/" + tableEscape(docID)
	if err := d.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &doc); err != nil {
		return contract.Document{}, err
	}
	return normalizeDocumentIDs(doc), nil
}

func (d *documentClient) Save(ctx context.Context, doc contract.Document) (contract.Document, error) {
	docID := preferredDocumentID(doc)
	if docID == "" {
		return contract.Document{}, fmt.Errorf("document id is required")
	}

	payload := doc
	if payload.DocumentID == "" {
		payload.DocumentID = docID
	}
	if payload.ID == "" {
		payload.ID = docID
	}

	path := d.basePath()
	var saved contract.Document
	if err := d.client.httpClient.DoJSON(ctx, http.MethodPut, path, payload, &saved); err != nil {
		return contract.Document{}, err
	}
	return normalizeDocumentIDs(saved), nil
}

func (d *documentClient) Delete(ctx context.Context, id string) error {
	docID := strings.TrimSpace(id)
	if docID == "" {
		return fmt.Errorf("document id is required")
	}

	path := d.basePath() + "/" + tableEscape(docID)
	return d.client.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}

func (d *documentClient) basePath() string {
	return "/data/" + tableEscape(d.client.cfg.DatabaseID) + "/document"
}

func preferredDocumentID(doc contract.Document) string {
	if id := strings.TrimSpace(doc.DocumentID); id != "" {
		return id
	}
	return strings.TrimSpace(doc.ID)
}

func normalizeDocumentIDs(doc contract.Document) contract.Document {
	if doc.DocumentID == "" {
		doc.DocumentID = strings.TrimSpace(doc.ID)
	}
	if doc.ID == "" {
		doc.ID = strings.TrimSpace(doc.DocumentID)
	}
	return doc
}
