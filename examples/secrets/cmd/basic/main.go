package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.Init(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	secretKey := fmt.Sprintf("example-secret-%d", time.Now().UnixMilli())
	secretValue := "demo-secret-value"

	entries, err := db.ListSecrets(ctx)
	if err != nil {
		log.Fatal(err)
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

	saved, err := db.PutSecret(ctx, contract.Secret{Key: secretKey, Value: secretValue})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saved secret metadata:", saved)

	afterSet, err := db.ListSecrets(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range afterSet {
		if s.Key == secretKey {
			fmt.Println("Secret now present in list:", s)
			break
		}
	}

	found, err := db.GetSecret(ctx, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Fetched secret:", found)

	if err := db.DeleteSecret(ctx, secretKey); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Secret deleted:", secretKey)

	finalList, err := db.ListSecrets(ctx)
	if err != nil {
		log.Fatal(err)
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
}
