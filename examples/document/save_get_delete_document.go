//go:build docs

package document

import (
	"context"
	"fmt"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// SaveGetDeleteDocument walks through the Documents API.
func SaveGetDeleteDocument(ctx context.Context) error {
	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		return err
	}
	docs := db.Documents()

	doc := contract.Document{
		ID: "doc_123",
		Data: map[string]any{
			"title":  "Welcome Guide",
			"status": "draft",
		},
	}

	saved, err := docs.Save(ctx, doc)
	if err != nil {
		return err
	}

	fmt.Printf("saved document: %+v\n", saved)

	found, err := docs.Get(ctx, saved.ID)
	if err != nil {
		return err
	}

	fmt.Printf("fetched document: %+v\n", found)

	if err := docs.Delete(ctx, saved.ID); err != nil {
		return err
	}

	fmt.Println("deleted document", saved.ID)
	return nil
}
