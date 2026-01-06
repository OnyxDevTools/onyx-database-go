package main

import (
	"context"
	"encoding/json"
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

	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	_, _ = db.Save(ctx, "AuditLog", map[string]any{
		"id":       "audit-id-a",
		"tenantId": "tenantA",
		"dateTime": now,
		"action":   "LOGIN",
		"status":   "SUCCESS",
	}, nil)

	_, _ = db.Save(ctx, "AuditLog", map[string]any{
		"id":       "audit-id-b",
		"tenantId": "tenantB",
		"dateTime": now,
		"action":   "LOGIN",
		"status":   "SUCCESS",
	}, nil)

	logs, err := db.From("AuditLog").
		Where(contract.Eq("tenantId", "tenantA")).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(logs, "", "  ")
	fmt.Println(string(out))
}
