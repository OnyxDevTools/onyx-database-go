package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
	coreonyx "github.com/OnyxDevTools/onyx-database-go/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	bucket := coreonyx.Format("dateTime", "yyyy-MM-dd HH")
	rows, err := db.AuditLogs().
		Select(bucket, coreonyx.Count("*"), "status").
		GroupBy(bucket, "status").
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if rows == nil {
		log.Fatalf("warning: expected bucketed rows")
	}

	out, _ := json.MarshalIndent(rows, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}
