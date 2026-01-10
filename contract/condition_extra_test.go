package contract

import (
	"encoding/json"
	"errors"
	"testing"
)

type errQuery struct{}

func (errQuery) MarshalJSON() ([]byte, error) { return nil, errors.New("boom") }

func TestConditionMarshalJSONBranches(t *testing.T) {
	// within with query marshal error
	c := condition{op: "within", field: "f", query: errQuery{}}
	if _, err := c.MarshalJSON(); err == nil {
		t.Fatalf("expected marshal error from nested query")
	}

	// between branch
	c = condition{op: "between", field: "f", from: 1, to: 2}
	data, err := c.MarshalJSON()
	if err != nil || !json.Valid(data) {
		t.Fatalf("marshal between err: %v", err)
	}

	// in with values nil should omit value
	c = condition{op: "in", field: "f"}
	data, err = c.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal in err: %v", err)
	}
	if string(data) == "" {
		t.Fatalf("expected payload")
	}

	// default operator branch
	if op := operatorFor("custom"); op != "custom" {
		t.Fatalf("expected passthrough operator, got %s", op)
	}
}

func TestErrorStringNilReceiver(t *testing.T) {
	var e *Error
	if e.Error() != "<nil>" {
		t.Fatalf("expected <nil> error string")
	}
}
