package onyx

import (
	"context"
	"testing"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type stubQuery struct {
	results contract.QueryResults
	err     error
}

func (s *stubQuery) Where(condition contract.Condition) contract.Query       { return s }
func (s *stubQuery) And(condition contract.Condition) contract.Query         { return s }
func (s *stubQuery) Or(condition contract.Condition) contract.Query          { return s }
func (s *stubQuery) Select(fields ...string) contract.Query                  { return s }
func (s *stubQuery) GroupBy(fields ...string) contract.Query                 { return s }
func (s *stubQuery) Resolve(paths ...string) contract.Query                  { return s }
func (s *stubQuery) OrderBy(sorts ...contract.Sort) contract.Query           { return s }
func (s *stubQuery) Limit(limit int) contract.Query                          { return s }
func (s *stubQuery) SetUpdates(updates map[string]any) contract.Query        { return s }
func (s *stubQuery) MarshalJSON() ([]byte, error)                            { return nil, nil }
func (s *stubQuery) List(ctx context.Context) (contract.QueryResults, error) { return s.results, s.err }
func (s *stubQuery) Page(ctx context.Context, cursor string) (contract.PageResult, error) {
	return contract.PageResult{}, nil
}
func (s *stubQuery) Stream(ctx context.Context) (contract.Iterator, error) { return nil, nil }
func (s *stubQuery) Update(ctx context.Context) (int, error)               { return 0, nil }
func (s *stubQuery) Delete(ctx context.Context) (int, error)               { return 0, nil }

func TestListIntoDecodesResults(t *testing.T) {
	q := &stubQuery{
		results: contract.QueryResults{
			map[string]any{"id": "1", "username": "alice"},
		},
	}

	var users []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	if err := ListInto(context.Background(), q, &users); err != nil {
		t.Fatalf("ListInto error: %v", err)
	}

	if len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("unexpected decoded users: %+v", users)
	}
}

func TestListFluentDecode(t *testing.T) {
	q := &stubQuery{
		results: contract.QueryResults{
			map[string]any{"id": "2", "username": "bob"},
		},
	}

	var users []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}

	if err := List(context.Background(), q).Decode(&users); err != nil {
		t.Fatalf("List Decode error: %v", err)
	}

	if len(users) != 1 || users[0].ID != "2" {
		t.Fatalf("unexpected decoded users: %+v", users)
	}
}

func TestListFluentPropagatesError(t *testing.T) {
	q := &stubQuery{err: context.Canceled}
	var users []struct{}
	if err := List(context.Background(), q).Decode(&users); err != context.Canceled {
		t.Fatalf("expected upstream error, got %v", err)
	}
}

func TestListIntoPropagatesError(t *testing.T) {
	q := &stubQuery{err: context.DeadlineExceeded}
	if err := ListInto(context.Background(), q, &[]struct{}{}); err != context.DeadlineExceeded {
		t.Fatalf("expected error propagation, got %v", err)
	}
}
