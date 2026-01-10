package contract

import "testing"

func TestCascadeBuilderBuildsSpec(t *testing.T) {
	spec := NewCascadeBuilder().
		Graph("roles.permissions").
		GraphType("Role").
		TargetField("id").
		SourceField("roleId").
		Build()
	if spec.String() == "" {
		t.Fatalf("expected cascade string")
	}
}

func TestCascadeBuilderGraphTypeOnly(t *testing.T) {
	spec := NewCascadeBuilder().
		GraphType("Role").
		Build()
	if spec.String() != "Role" {
		t.Fatalf("expected graph type only cascade, got %q", spec.String())
	}
}
