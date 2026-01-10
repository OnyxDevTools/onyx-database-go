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
