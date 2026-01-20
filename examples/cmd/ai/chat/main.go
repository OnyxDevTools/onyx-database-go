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

	resp, err := db.Chat(ctx, onyx.AIChatCompletionRequest{
		Model: "onyx-chat",
		Messages: []onyx.AIChatMessage{
			{Role: "user", Content: "Reply with one short sentence (under 12 words) saying hello and mentioning Onyx."},
		},
	})
	if err != nil {
		log.Fatalf("chat call failed: %v", err)
	}
	if resp.ID == "" || len(resp.Choices) == 0 {
		log.Fatalf("chat response missing data: %+v", resp)
	}
	if resp.Choices[0].Message.Content == "" {
		log.Fatalf("chat content empty")
	}

	log.Println("example: completed")
}
