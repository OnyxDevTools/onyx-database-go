package onyx

import (
	"context"
	"fmt"
	"net/http"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type documentClient struct {
	client *client
}

func (c *client) Documents() contract.DocumentClient {
	return &documentClient{client: c}
}

func (d *documentClient) List(ctx context.Context) ([]contract.Document, error) {
	var docs []contract.Document
	if err := d.client.httpClient.DoJSON(ctx, http.MethodGet, "/documents", nil, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (d *documentClient) Get(ctx context.Context, id string) (contract.Document, error) {
	if id == "" {
		return contract.Document{}, fmt.Errorf("document id is required")
	}
	var doc contract.Document
	path := "/documents/" + id
	if err := d.client.httpClient.DoJSON(ctx, http.MethodGet, path, nil, &doc); err != nil {
		return contract.Document{}, err
	}
	return doc, nil
}

func (d *documentClient) Save(ctx context.Context, doc contract.Document) (contract.Document, error) {
	if doc.ID == "" {
		return contract.Document{}, fmt.Errorf("document id is required")
	}
	path := "/documents/" + doc.ID
	var saved contract.Document
	if err := d.client.httpClient.DoJSON(ctx, http.MethodPut, path, doc, &saved); err != nil {
		return contract.Document{}, err
	}
	return saved, nil
}

func (d *documentClient) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("document id is required")
	}
	path := "/documents/" + id
	return d.client.httpClient.DoJSON(ctx, http.MethodDelete, path, nil, nil)
}
