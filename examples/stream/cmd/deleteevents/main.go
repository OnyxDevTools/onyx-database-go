package main

import (
	"context"
	"log"
	"time"

	model "github.com/OnyxDevTools/onyx-database-go/examples/onyx"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	streamDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}
	writeDB, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	iter, err := streamDB.From(model.Tables.User).Stream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	// seed then delete to trigger an event
	go func() {
		time.Sleep(200 * time.Millisecond)
		now := time.Now().UTC()
		_, _ = writeDB.Save(ctx, model.Tables.User, model.User{
			Id:        "stream_user_delete",
			Username:  "delete-user",
			Email:     "delete@example.com",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)
		time.Sleep(200 * time.Millisecond)
		_ = writeDB.Delete(ctx, model.Tables.User, "stream_user_delete")
	}()

	for iter.Next() {
		log.Printf("USER EVENT: %+v", iter.Value())
	}
	if err := iter.Err(); err != nil {
		log.Fatal(err)
	}
}
