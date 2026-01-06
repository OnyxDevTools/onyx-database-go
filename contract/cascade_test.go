package contract

import "testing"

func TestCascadeString(t *testing.T) {
	spec := Cascade("userRoles:UserRole(userId,id)")
	if got := spec.String(); got != "userRoles:UserRole(userId,id)" {
		t.Fatalf("unexpected spec string: %s", got)
	}
}

func TestCascadeBuilder(t *testing.T) {
	spec := NewCascadeBuilder().
		Graph("userRoles").
		GraphType("UserRole").
		SourceField("userId").
		TargetField("id").
		Build()

	if got := spec.String(); got != "userRoles:UserRole(userId,id)" {
		t.Fatalf("unexpected builder output: %s", got)
	}
}
