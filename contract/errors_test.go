package contract

import "testing"

func TestErrorStringFormatting(t *testing.T) {
	err := &Error{
		Code:    "ERR_TEST",
		Message: "something went wrong",
		Meta: map[string]any{
			"beta":  2,
			"alpha": 1,
		},
	}

	got := err.Error()
	want := "ERR_TEST: something went wrong [alpha=1, beta=2]"

	if got != want {
		t.Fatalf("unexpected error string. got=%q want=%q", got, want)
	}
}

func TestErrorNilMeta(t *testing.T) {
	err := &Error{Code: "ERR_TEST", Message: "nil meta"}

	if got := err.Error(); got != "ERR_TEST: nil meta" {
		t.Fatalf("unexpected error string for nil meta: %q", got)
	}
}

func TestNewErrorHelper(t *testing.T) {
	meta := map[string]any{"info": "details"}
	err := NewError("ERR_HELPER", "constructed", meta)

	if err.Code != "ERR_HELPER" || err.Message != "constructed" || err.Meta["info"] != "details" {
		t.Fatalf("unexpected error fields: %#v", err)
	}
}
