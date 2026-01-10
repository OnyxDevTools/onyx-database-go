package generator

import "testing"

func TestToExportedFallbacks(t *testing.T) {
	if got := toExported("###"); got != "Name" {
		t.Fatalf("expected Name fallback, got %s", got)
	}
	if got := toExported("snake_case-name"); got != "SnakeCaseName" {
		t.Fatalf("unexpected toExported conversion: %s", got)
	}
}
