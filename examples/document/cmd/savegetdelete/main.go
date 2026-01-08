package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyxclient"
)

func main() {
	ctx := context.Background()

	core, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db := onyxclient.NewClient(core)
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
	fmt.Printf("saved document: %+v\n", saved)

	found, err := docs.Get(ctx, saved.ID)
	if err != nil {
		log.Fatal(err)
	}
	if found.Content != "" {
		if decoded, decodeErr := base64.StdEncoding.DecodeString(found.Content); decodeErr == nil {
			fmt.Printf("Document contents: %s\n", decoded)
		}
	}

	if err := docs.Delete(ctx, saved.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted document:", saved.ID)

	if _, err := docs.Get(ctx, saved.ID); err != nil {
		fmt.Println("Document was deleted successfully")
	} else {
		fmt.Println("oops document still exists")
	}
}
