package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	maxTokens := 32
	stream, err := db.ChatStream(streamCtx, onyx.AIChatCompletionRequest{
		Model:      "onyx-chat",
		MaxTokens:  &maxTokens,
		ToolChoice: "none",
		Messages: []onyx.AIChatMessage{
			{
				Role:    "user",
				Content: "Return today's date as YYYY-MM-DD in one short sentence (<=12 words). Do not call tools or include links.",
			},
		},
	})
	if err != nil {
		log.Fatalf("chat stream failed: %v", err)
	}
	defer stream.Close()

	var content strings.Builder
	for stream.Next() {
		chunk := stream.Chunk()
		if chunk.ID == "" {
			log.Fatalf("stream chunk missing id")
		}
		if len(chunk.Choices) > 0 {
			content.WriteString(chunk.Choices[0].Delta.Content)
		}
	}
	if err := stream.Err(); err != nil {
		log.Fatalf("stream error: %v", err)
	}
	if content.Len() == 0 {
		log.Fatalf("stream content was empty")
	}

	log.Println("example: completed")
}
