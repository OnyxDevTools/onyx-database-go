//go:build docs

package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/onyx"
)

// readmeInitExamples mirrors the README snippets and is compiled under the "docs" build tag.
func readmeInitExamples(ctx context.Context) {
	_, _ = onyx.Init(ctx, onyx.Config{
		DatabaseID:      "db_123",
		DatabaseBaseURL: "https://api.onyx.dev",
		APIKey:          os.Getenv("ONYX_DATABASE_API_KEY"),
		APISecret:       os.Getenv("ONYX_DATABASE_API_SECRET"),
		LogRequests:     true,
		LogResponses:    false,
	})

	_, _ = onyx.InitWithDatabaseID(ctx, "db_123")

	os.Setenv("ONYX_DATABASE_ID", "db_123")
	os.Setenv("ONYX_DATABASE_BASE_URL", "https://api.onyx.dev")
	os.Setenv("ONYX_DATABASE_API_KEY", "key_abc")
	os.Setenv("ONYX_DATABASE_API_SECRET", "secret_xyz")

	defer onyx.ClearConfigCache()
}

func readmeQueryExamples(client contract.Client, ctx context.Context) {
	page, _ := client.From("User").
		Select("id", "email", "profile.avatarUrl").
		Where(contract.Contains("email", "@example.com")).
		And(contract.Gte("createdAt", "2024-01-01T00:00:00Z")).
		Resolve("roles", "profile").
		OrderBy(contract.Desc("createdAt")).
		Limit(25).
		Page(ctx, "")
	_ = page

	_, _ = client.From("Event").Stream(ctx)

	_, _ = client.From("Post").
		Where(contract.Eq("id", "post_123")).
		Resolve("author.profile", "comments.author").
		Limit(1).
		List(ctx)

	spec := contract.Cascade("userRoles:UserRole(userId,id)")
	fmt.Println(spec.String())

	_ = client.Cascade(spec)
}
