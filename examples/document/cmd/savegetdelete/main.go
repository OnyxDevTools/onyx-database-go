package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := core.Typed()
	docs := db.Documents()

	payload := map[string]any{"message": "hello"}
	encoded, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}

	doc := onyx.Document{
		DocumentID: "note.json",
		Path:       "/notes/note.json",
		MimeType:   "application/json",
		Content:    base64.StdEncoding.EncodeToString(encoded),
	}

	saved, err := docs.Save(ctx, doc)
	if err != nil {
		log.Fatal(err)
	}
	if saved.ID == "" {
		log.Fatalf("warning: expected saved document id")
	}
	fmt.Printf("saved document: %+v\n", saved)

	_, err = docs.Get(ctx, doc.DocumentID)
	if err != nil {
		log.Fatal(err)
	}

	if err := docs.Delete(ctx, doc.DocumentID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted document:", doc.DocumentID)

	if _, err := docs.Get(ctx, doc.DocumentID); err != nil {
		fmt.Println("Document was deleted successfully")
	} else {
		log.Fatalf("warning: expected document to be deleted")
	}
	log.Println("example: completed")
}
