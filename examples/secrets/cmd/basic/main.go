package main

import (
	"context"
	"fmt"
	"log"
	"time"

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
	secrets := db.Secrets()

	secretKey := fmt.Sprintf("example-secret-%d", time.Now().UnixMilli())
	secretValue := "demo-secret-value"

	entries, err := secrets.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if entries == nil {
		log.Println("warning: expected secrets list")
	}
	exists := false
	for _, s := range entries {
		if s.Key == secretKey {
			exists = true
		}
	}
	if !exists {
		fmt.Println("Initial secrets list does not include the example key.")
	}

	saved, err := secrets.Set(ctx, onyx.Secret{Key: secretKey, Value: secretValue})
	if err != nil {
		log.Fatal(err)
	}
	if saved.Key != secretKey {
		log.Println("warning: expected saved secret key")
	}
	fmt.Println("Saved secret metadata:", saved)

	afterSet, err := secrets.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if afterSet == nil {
		log.Println("warning: expected secrets list after set")
	}
	for _, s := range afterSet {
		if s.Key == secretKey {
			fmt.Println("Secret now present in list:", s)
			break
		}
	}

	found, err := secrets.Get(ctx, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	if found.Key != secretKey {
		log.Println("warning: expected fetched secret key")
	}
	fmt.Println("Fetched secret:", found)

	if err := secrets.Delete(ctx, secretKey); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Secret deleted:", secretKey)

	finalList, err := secrets.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if finalList == nil {
		log.Println("warning: expected final secrets list")
	}
	stillExists := false
	for _, s := range finalList {
		if s.Key == secretKey {
			stillExists = true
		}
	}
	if !stillExists {
		fmt.Println("Final secrets list confirms removal.")
	}
	if stillExists {
		log.Println("warning: expected secret to be removed")
	}
	log.Println("example: completed")
}
