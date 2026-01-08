package contract

import "testing"

func TestQueryResultsDecode(t *testing.T) {
	results := QueryResults{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}

	var users []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := results.Decode(&users); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if len(users) != 2 || users[0].Name != "Alice" || users[1].ID != "2" {
		t.Fatalf("unexpected decoded users: %+v", users)
	}
}
