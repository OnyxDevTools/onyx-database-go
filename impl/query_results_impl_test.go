package impl

import "testing"

func TestToQueryResultsNilAndValues(t *testing.T) {
	if res := toQueryResults(nil); res == nil || len(res) != 0 {
		t.Fatalf("expected empty slice for nil input")
	}

	items := []map[string]any{{"id": 1}}
	res := toQueryResults(items)
	if len(res) != 1 || res[0]["id"] != 1 {
		t.Fatalf("unexpected results: %+v", res)
	}
}
