package contract

import "testing"

func TestPageResultUnmarshalErrors(t *testing.T) {
	var p PageResult
	if err := p.UnmarshalJSON([]byte(`{"records":[{"id":1}],"nextPage":"c"}`)); err != nil {
		t.Fatalf("expected alt shape decode, got %v", err)
	}
	if err := p.UnmarshalJSON([]byte("{")); err == nil {
		t.Fatalf("expected error on bad json")
	}
}

func TestQueryResultsUnmarshalErrors(t *testing.T) {
	var q QueryResults
	if err := q.UnmarshalJSON([]byte("{")); err == nil {
		t.Fatalf("expected error on bad json")
	}
}
