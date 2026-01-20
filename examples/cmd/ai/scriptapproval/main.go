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

	resp, err := db.RequestScriptApproval(ctx, onyx.AIScriptApprovalRequest{
		Script: "db.save({ id: 123, name: \"ai-example\" })",
	})
	if err != nil {
		log.Fatalf("script approval failed: %v", err)
	}
	if resp.NormalizedScript == "" {
		log.Fatalf("normalized script missing")
	}
	if resp.ExpiresAtIso == "" {
		log.Fatalf("expiresAtIso missing")
	}

	log.Println("example: completed")
}
