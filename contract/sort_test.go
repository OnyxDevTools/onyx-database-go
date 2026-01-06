package contract

import (
	"encoding/json"
	"testing"
)

func TestSortJSON(t *testing.T) {
	cases := []struct {
		name string
		sort Sort
		want string
	}{
		{name: "asc", sort: Asc("createdAt"), want: `{"field":"createdAt","direction":"asc"}`},
		{name: "desc", sort: Desc("name"), want: `{"field":"name","direction":"desc"}`},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.sort)
			if err != nil {
				t.Fatalf("marshal sort: %v", err)
			}

			if string(data) != tt.want {
				t.Fatalf("unexpected json. got=%s want=%s", string(data), tt.want)
			}
		})
	}
}
