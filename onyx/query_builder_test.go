package onyx

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
	expected := `{"table":"users","where":{"and":[{"field":"name","op":"eq","value":"a"},{"field":"age","op":"gt","value":1}]},"orderBy":[{"field":"name","direction":"asc"}]}`
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
