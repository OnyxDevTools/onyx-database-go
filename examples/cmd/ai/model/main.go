package main

import (
	"context"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	models, err := db.GetModels(ctx)
	if err != nil {
		log.Fatalf("list models failed: %v", err)
	}
	if len(models.Data) == 0 {
		log.Fatalf("no models available to fetch")
	}
	target := models.Data[0].ID

	model, err := db.GetModel(ctx, target)
	if err != nil {
		log.Fatalf("get model failed: %v", err)
	}
	if model.ID == "" || model.ID != target {
		log.Fatalf("unexpected model payload: %+v", model)
	}

	log.Println("example: completed")
}
