package impl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

const fullTextSearchField = "__full_text__"

type conditionPlan struct {
	candidateOperator string
	readOnlyOperator  string
	searchCount       int
	searchWrongField  bool
	fullTextCount     int
	rootIsSingle      bool
	rootOperator      string
}

// inspectConditionOperators avoids an extra marshal for SDK-owned conditions,
// while still recognizing operators hidden inside arbitrary public condition
// marshalers such as json.RawMessage.
func inspectConditionOperators(condition contract.Condition) (conditionPlan, error) {
	type candidateCondition interface {
		CandidateOperator() string
	}
	type readOnlyCondition interface {
		ReadOnlyOperator() string
	}

	candidate, candidateKnown := condition.(candidateCondition)
	readOnly, readOnlyKnown := condition.(readOnlyCondition)
	if candidateKnown && readOnlyKnown {
		candidateOperator := canonicalCandidateOperator(candidate.CandidateOperator())
		return conditionPlan{
			candidateOperator: candidateOperator,
			readOnlyOperator:  canonicalReadOnlyOperator(readOnly.ReadOnlyOperator()),
			rootIsSingle:      true,
			rootOperator:      candidateOperator,
		}, nil
	}

	raw, err := json.Marshal(condition)
	if err != nil {
		return conditionPlan{}, fmt.Errorf("marshal query condition: %w", err)
	}
	return inspectConditionJSON(raw)
}

func inspectConditionJSON(raw json.RawMessage) (conditionPlan, error) {
	if len(raw) == 0 {
		return conditionPlan{}, nil
	}

	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return conditionPlan{}, fmt.Errorf("decode query condition: %w", err)
	}

	plan := conditionPlan{}
	rootNode, rootIsObject := root.(map[string]any)
	if rootIsObject && isConditionType(rootNode, "SingleCondition") {
		plan.rootIsSingle = true
		plan.rootOperator, _ = criteriaOperatorAndField(rootNode)
	}
	inspectConditionNode(root, &plan)
	return plan, nil
}

func inspectConditionNode(value any, plan *conditionPlan) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}

	switch {
	case isConditionType(node, "SingleCondition"):
		operator, field := criteriaOperatorAndField(node)
		if candidate := canonicalCandidateOperator(operator); candidate != "" && plan.candidateOperator == "" {
			plan.candidateOperator = candidate
		}
		if readOnly := canonicalReadOnlyOperator(operator); readOnly != "" && plan.readOnlyOperator == "" {
			plan.readOnlyOperator = readOnly
		}
		if canonicalReadOnlyOperator(operator) == "SEARCH" {
			plan.searchCount++
			if field != fullTextSearchField {
				plan.searchWrongField = true
			}
		}
		if field == fullTextSearchField {
			plan.fullTextCount++
		}
	case isConditionType(node, "CompoundCondition"):
		conditions, ok := node["conditions"].([]any)
		if !ok {
			return
		}
		for _, condition := range conditions {
			inspectConditionNode(condition, plan)
		}
	}
}

func validateConditionPlan(raw json.RawMessage) error {
	plan, err := inspectConditionJSON(raw)
	if err != nil {
		return err
	}

	if plan.candidateOperator != "" &&
		(!plan.rootIsSingle || plan.rootOperator != plan.candidateOperator) {
		return fmt.Errorf("%s must be the sole root criterion", plan.candidateOperator)
	}
	if plan.searchCount > 1 {
		return fmt.Errorf("a query may contain only one SEARCH criterion")
	}
	if plan.searchCount == 1 {
		if plan.searchWrongField {
			return fmt.Errorf("SEARCH requires the %s field", fullTextSearchField)
		}
		if plan.fullTextCount > 1 {
			return fmt.Errorf("SEARCH cannot be combined with another full-text search criterion")
		}
	}
	return nil
}

func isConditionType(node map[string]any, want string) bool {
	conditionType, _ := node["conditionType"].(string)
	return strings.EqualFold(strings.TrimSpace(conditionType), want)
}

func criteriaOperatorAndField(node map[string]any) (string, string) {
	criteria, ok := node["criteria"].(map[string]any)
	if !ok {
		return "", ""
	}
	operator, _ := criteria["operator"].(string)
	field, _ := criteria["field"].(string)
	return strings.ToUpper(strings.TrimSpace(operator)), strings.TrimSpace(field)
}

func canonicalCandidateOperator(operator string) string {
	switch strings.ToUpper(strings.TrimSpace(operator)) {
	case "CANDIDATES":
		return "CANDIDATES"
	case "SEARCH_CANDIDATES":
		return "SEARCH_CANDIDATES"
	case "HNSW_CANDIDATES":
		return "HNSW_CANDIDATES"
	default:
		return ""
	}
}

func canonicalReadOnlyOperator(operator string) string {
	if candidate := canonicalCandidateOperator(operator); candidate != "" {
		return candidate
	}
	if strings.EqualFold(strings.TrimSpace(operator), "SEARCH") {
		return "SEARCH"
	}
	return ""
}
