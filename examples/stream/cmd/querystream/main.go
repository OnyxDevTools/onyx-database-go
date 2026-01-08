package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()
	streamCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := db.From("User").Stream(streamCtx)
	if err != nil {
		log.Fatal(err)
	}
	if iter == nil {
		log.Println("warning: expected stream iterator")
		return
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Printf("stream close error: %v", err)
		}
	}()

	count := 0
	for iter.Next() {
		user := iter.Value()
		if user == nil {
			log.Println("warning: expected streamed user")
		}
		fmt.Println("USER:", user)
		count++
		if count >= 3 {
			break
		}
	}

	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
	log.Println("example: completed")
}
