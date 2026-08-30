package impl

import (
	"fmt"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

type clause struct {
	Type      string
	Condition contract.Condition
}

type query struct {
	client        *client
	table         string
	clauses       []clause
	selectFields  []string
	groupFields   []string
	distinct      bool
	resolveFields []string
	sorts         []contract.Sort
	limit         *int
	updates       map[string]any
	partition     *string
	candidateOp   string
	readOnlyOp    string
	buildErr      error
}

func newQuery(client *client, table string) contract.Query {
	var partition *string
	if client != nil && table != "ALL" && strings.TrimSpace(client.cfg.Partition) != "" {
		p := strings.TrimSpace(client.cfg.Partition)
		partition = &p
	}
	return &query{client: client, table: table, partition: partition}
}

func (q *query) clone() *query {
	nq := *q
	nq.clauses = append([]clause{}, q.clauses...)
	nq.selectFields = append([]string{}, q.selectFields...)
	nq.groupFields = append([]string{}, q.groupFields...)
	nq.resolveFields = append([]string{}, q.resolveFields...)
	nq.sorts = append([]contract.Sort{}, q.sorts...)
	if q.partition != nil {
		p := *q.partition
		nq.partition = &p
	}
	if q.updates != nil {
		nq.updates = map[string]any{}
		for k, v := range q.updates {
			nq.updates[k] = v
		}
	}
	return &nq
}

func (q *query) Where(condition contract.Condition) contract.Query {
	nq := q.clone()
	if nq.buildErr != nil {
		return nq
	}
	if nq.candidateOp != "" {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", nq.candidateOp)
		return nq
	}
	plan, err := inspectConditionOperators(condition)
	if err != nil {
		nq.buildErr = err
		return nq
	}
	incomingCandidate := plan.candidateOperator
	incomingReadOnly := plan.readOnlyOperator
	if incomingReadOnly != "" {
		nq.readOnlyOp = incomingReadOnly
	}
	if incomingCandidate != "" &&
		(!plan.rootIsSingle || plan.rootOperator != incomingCandidate) {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", incomingCandidate)
		return nq
	}
	if incomingCandidate != "" && len(nq.clauses) > 0 {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", incomingCandidate)
		return nq
	}
	nq.clauses = append(nq.clauses, clause{Type: "and", Condition: condition})
	nq.candidateOp = incomingCandidate
	return nq
}

func (q *query) And(condition contract.Condition) contract.Query {
	return q.Where(condition)
}

func (q *query) Or(condition contract.Condition) contract.Query {
	nq := q.clone()
	if nq.buildErr != nil {
		return nq
	}
	if nq.candidateOp != "" {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", nq.candidateOp)
		return nq
	}
	plan, err := inspectConditionOperators(condition)
	if err != nil {
		nq.buildErr = err
		return nq
	}
	incomingCandidate := plan.candidateOperator
	incomingReadOnly := plan.readOnlyOperator
	if incomingReadOnly != "" {
		nq.readOnlyOp = incomingReadOnly
	}
	if incomingCandidate != "" &&
		(!plan.rootIsSingle || plan.rootOperator != incomingCandidate) {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", incomingCandidate)
		return nq
	}
	if incomingCandidate != "" && len(nq.clauses) > 0 {
		nq.buildErr = fmt.Errorf("%s must be the sole root criterion", incomingCandidate)
		return nq
	}
	nq.clauses = append(nq.clauses, clause{Type: "or", Condition: condition})
	nq.candidateOp = incomingCandidate
	return nq
}

func (q *query) Search(queryText string, minScore ...float64) contract.Query {
	return q.Where(contract.Search(queryText, minScore...))
}

func (q *query) SearchWithOptions(
	queryText string,
	options contract.SearchOptions,
) contract.Query {
	return q.Where(contract.SearchWithOptions(queryText, options))
}

func (q *query) SearchVector(searchQuery contract.VectorSearchQuery) contract.Query {
	return q.Where(contract.VectorSearch(searchQuery))
}

func (q *query) ApproximateSearch(searchQuery contract.VectorSearchQuery) contract.Query {
	return q.Where(contract.ApproximateSearch(searchQuery))
}

func (q *query) HNSWCandidates(searchQuery contract.HNSWSearchQuery) contract.Query {
	return q.Where(contract.HNSWCandidates(searchQuery))
}

func (q *query) ApproximateCandidates(
	attribute string,
	valueOrValues any,
	maxCandidates ...int,
) contract.Query {
	return q.Where(contract.ApproximateCandidates(attribute, valueOrValues, maxCandidates...))
}

func (q *query) Select(fields ...string) contract.Query {
	nq := q.clone()
	nq.selectFields = append(nq.selectFields, fields...)
	return nq
}

func (q *query) GroupBy(fields ...string) contract.Query {
	nq := q.clone()
	nq.groupFields = append(nq.groupFields, fields...)
	return nq
}

func (q *query) Distinct() contract.Query {
	nq := q.clone()
	nq.distinct = true
	return nq
}

func (q *query) Resolve(paths ...string) contract.Query {
	nq := q.clone()
	nq.resolveFields = append(nq.resolveFields, paths...)
	return nq
}

func (q *query) OrderBy(sorts ...contract.Sort) contract.Query {
	nq := q.clone()
	nq.sorts = append(nq.sorts, sorts...)
	return nq
}

func (q *query) Limit(limit int) contract.Query {
	nq := q.clone()
	nq.limit = &limit
	return nq
}

func (q *query) SetUpdates(updates map[string]any) contract.Query {
	nq := q.clone()
	nq.updates = map[string]any{}
	for k, v := range updates {
		nq.updates[k] = v
	}
	return nq
}

func (q *query) InPartition(partition string) contract.Query {
	nq := q.clone()
	trimmed := strings.TrimSpace(partition)
	if trimmed == "" {
		nq.partition = nil
		return nq
	}
	nq.partition = &trimmed
	return nq
}

func (q *query) MarshalJSON() ([]byte, error) {
	payload, err := buildQueryPayload(q, true)
	if err != nil {
		return nil, err
	}
	return payload.MarshalJSON()
}
