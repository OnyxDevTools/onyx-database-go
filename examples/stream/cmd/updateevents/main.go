package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	streamCore, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeCore, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	streamDB := streamCore.Typed()
	writeDB := writeCore.Typed()

	iter, err := streamDB.ListUsers().Stream(streamCtx)
	if err != nil {
		log.Fatal(err)
	}
	if iter == nil {
		log.Fatalf("warning: expected stream iterator")
		return
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Printf("stream close error: %v", err)
		}
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, err := writeDB.SaveUser(ctx, onyx.User{
			Id:        "stream_user_update",
			Username:  "update-user",
			Email:     "update@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			log.Printf("save error: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
		updated := time.Now().UTC()
		lastLogin := updated
		_, err = writeDB.SaveUser(ctx, onyx.User{
			Id:          "stream_user_update",
			Username:    "update-user-updated",
			Email:       "update@example.com",
			IsActive:    true,
			LastLoginAt: &lastLogin,
			CreatedAt:   updated,
			UpdatedAt:   updated,
		})
		if err != nil {
			log.Printf("save error: %v", err)
		}
	}()

	eventCount := 0
	for iter.Next() {
		if iter.Value() == nil {
			log.Fatalf("warning: expected streamed value")
		}
		fmt.Printf("USER EVENT: %+v\n", iter.Value())
		eventCount++
		if eventCount >= 2 {
			break
		}
	}
	if err := iter.Err(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
