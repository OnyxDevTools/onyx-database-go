package impl

import (
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestQueryBuilderCloneMethods(t *testing.T) {
	q := &query{table: "users"}
	q2 := q.Or(contract.Eq("id", 1)).GroupBy("role").Resolve("profile").Limit(5).SetUpdates(map[string]any{"email": "x"})

	cq, ok := q2.(*query)
	if !ok {
		t.Fatalf("expected *query")
	}
	if len(cq.groupFields) != 1 || cq.limit == nil || *cq.limit != 5 {
		t.Fatalf("clone did not apply operations: %+v", cq)
	}
	if cq.updates["email"] != "x" {
		t.Fatalf("updates not set")
	}
}
