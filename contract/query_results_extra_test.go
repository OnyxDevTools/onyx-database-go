package contract

import "testing"

func TestQueryResultsUnmarshalWrapper(t *testing.T) {
	raw := []byte(`{"records":[{"id":"1"}]}`)
	var qr QueryResults
	if err := qr.UnmarshalJSON(raw); err != nil {
		t.Fatalf("unmarshal wrapper: %v", err)
	}
	if len(qr) != 1 || qr[0]["id"] != "1" {
		t.Fatalf("unexpected results: %+v", qr)
	}
}
