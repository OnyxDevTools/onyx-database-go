package impl

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
	"github.com/OnyxDevTools/onyx-database-go/impl/resolver"
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

func TestSearchPayloads(t *testing.T) {
	client := &client{cfg: resolver.ResolvedConfig{DatabaseID: "db"}}
	cases := []struct {
		name     string
		q        contract.Query
		expected string
	}{
		{
			name:     "table search with min score",
			q:        newQuery(nil, "Table").Search("Text", 4.4),
			expected: `{"type":"SelectQuery","table":"Table","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"minScore":4.4,"queryText":"Text"}}}}`,
		},
		{
			name:     "table search null min score",
			q:        newQuery(nil, "Table").Search("Text"),
			expected: `{"type":"SelectQuery","table":"Table","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"minScore":null,"queryText":"Text"}}}}`,
		},
		{
			name:     "all tables search",
			q:        client.Search("Text", 4.4),
			expected: `{"type":"SelectQuery","table":"ALL","conditions":{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"minScore":4.4,"queryText":"Text"}}}}`,
		},
		{
			name: "search combined with filter",
			q: newQuery(nil, "Table").
				Search("text", 4.4).
				And(contract.Eq("attribute", "value")),
			expected: `{"type":"SelectQuery","table":"Table","conditions":{"conditionType":"CompoundCondition","conditions":[{"conditionType":"SingleCondition","criteria":{"field":"__full_text__","operator":"MATCHES","value":{"minScore":4.4,"queryText":"text"}}},{"conditionType":"SingleCondition","criteria":{"field":"attribute","operator":"EQUAL","value":"value"}}],"operator":"AND"}}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.q)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			if string(data) != tt.expected {
				t.Fatalf("unexpected payload:\n%s", string(data))
			}
		})
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

func TestPartitionDefaultsAndOverrides(t *testing.T) {
	client := &client{cfg: resolver.ResolvedConfig{DatabaseID: "db", Partition: "default"}}
	q := newQuery(client, "users")
	payload := buildQueryPayload(q.(*query), true)
	if payload.Partition == nil || *payload.Partition != "default" {
		t.Fatalf("expected default partition applied, got %+v", payload.Partition)
	}

	q2 := q.InPartition("p1").(*query)
	payload2 := buildQueryPayload(q2, true)
	if payload2.Partition == nil || *payload2.Partition != "p1" {
		t.Fatalf("expected override partition, got %+v", payload2.Partition)
	}

	q3 := q2.InPartition(" ").(*query)
	if q3.partition != nil {
		t.Fatalf("expected clearing partition when empty")
	}
}
