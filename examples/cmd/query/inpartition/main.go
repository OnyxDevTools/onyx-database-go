package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/OnyxDevTools/onyx-database-go/examples/gen/onyx"
)

func main() {
	ctx := context.Background()

	db, err := onyx.New(ctx, onyx.Config{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC()
	_, err = db.AuditLogs().Save(ctx, onyx.AuditLog{
		Id:       "audit-id-a",
		TenantId: strPtr("tenantA"),
		DateTime: now,
		Action:   strPtr("LOGIN"),
		Status:   strPtr("SUCCESS"),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.AuditLogs().Save(ctx, onyx.AuditLog{
		Id:       "audit-id-b",
		TenantId: strPtr("tenantB"),
		DateTime: now,
		Action:   strPtr("LOGIN"),
		Status:   strPtr("SUCCESS"),
	})
	if err != nil {
		log.Fatal(err)
	}

	logs, err := db.AuditLogs().
		Where(onyx.Eq("tenantId", "tenantA")).
		List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if logs == nil {
		log.Fatalf("warning: expected audit log response")
	}

	out, _ := json.MarshalIndent(logs, "", "  ")
	fmt.Println(string(out))
	log.Println("example: completed")
}

func strPtr(s string) *string { return &s }
