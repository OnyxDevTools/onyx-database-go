package impl

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

func TestQueryChainingImmutable(t *testing.T) {
	base := newQuery(nil, "users")
	q1 := base.Where(contract.Eq("name", "a"))
	q2 := base.Select("id")

	if reflect.DeepEqual(q1, q2) {
		t.Fatalf("queries should differ")
	}
}

func TestQueryMarshalDeterministic(t *testing.T) {
	q := newQuery(nil, "users").Where(contract.Eq("name", "a")).And(contract.Gt("age", 1)).OrderBy(contract.Asc("name"))
	data, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	expected := `{"type":"SelectQuery","table":"users","conditions":{"conditionType":"CompoundCondition","conditions":[{"conditionType":"SingleCondition","criteria":{"field":"name","operator":"EQUAL","value":"a"}},{"conditionType":"SingleCondition","criteria":{"field":"age","operator":"GREATER_THAN","value":1}}],"operator":"AND"},"sort":[{"field":"name","direction":"asc"}]}`
	if string(data) != expected {
		t.Fatalf("unexpected payload: %s", string(data))
	}
}

func TestWithinConditionEmbedsQuery(t *testing.T) {
	sub := newQuery(nil, "pets").Where(contract.Eq("type", "cat"))
	q := newQuery(nil, "users").Where(contract.Within("pet", sub))
	data, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), "\"table\":\"pets\"") {
		t.Fatalf("expected nested query: %s", data)
	}
}
