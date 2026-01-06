package contract

import (
	"fmt"
	"sort"
	"strings"
)

// Error represents a structured error returned by the SDK and CLI.
type Error struct {
	Code    string
	Message string
	Meta    map[string]any
}

// Error implements the error interface with deterministic formatting.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}

	base := e.Message
	if e.Code != "" {
		base = fmt.Sprintf("%s: %s", e.Code, e.Message)
	}

	if len(e.Meta) == 0 {
		return base
	}

	keys := make([]string, 0, len(e.Meta))
	for key := range e.Meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	metaPairs := make([]string, 0, len(keys))
	for _, key := range keys {
		metaPairs = append(metaPairs, fmt.Sprintf("%s=%v", key, e.Meta[key]))
	}

	return fmt.Sprintf("%s [%s]", base, strings.Join(metaPairs, ", "))
}

// NewError constructs a structured error with the provided fields.
func NewError(code, message string, meta map[string]any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Meta:    meta,
	}
}
