package impl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OnyxDevTools/onyx-database-go/contract"
)

const fullTextSearchField = "__full_text__"

type conditionPlan struct {
	soleRootCandidateOperator string
	approximateCandidateCount int
	readOnlyOperator          string
	searchCount               int
	searchWrongField          bool
	fullTextCount             int
	rootIsSingle              bool
	rootOperator              string
	pureConjunction           bool
}

// inspectConditionOperators avoids an extra marshal for SDK-owned single
// conditions while still recognizing operators and logical structure inside
// arbitrary public condition marshalers such as json.RawMessage.
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
		readOnlyOperator := canonicalReadOnlyOperator(readOnly.ReadOnlyOperator())
		plan := conditionPlan{
			readOnlyOperator: readOnlyOperator,
			rootIsSingle:     true,
			rootOperator:     candidateOperator,
			pureConjunction:  true,
		}
		switch candidateOperator {
		case "CANDIDATES":
			plan.approximateCandidateCount = 1
		case "SEARCH_CANDIDATES", "HNSW_CANDIDATES":
			plan.soleRootCandidateOperator = candidateOperator
			plan.fullTextCount = 1
		}
		if readOnlyOperator == "SEARCH" {
			plan.searchCount = 1
			plan.fullTextCount = 1
		}
		return plan, nil
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
	plan.pureConjunction = inspectConditionNode(root, &plan)
	return plan, nil
}

func inspectConditionNode(value any, plan *conditionPlan) bool {
	node, ok := value.(map[string]any)
	if !ok {
		return false
	}
	nonNegated := !conditionNodeIsNegated(node)

	switch {
	case isConditionType(node, "SingleCondition"):
		operator, field := criteriaOperatorAndField(node)
		if candidate := canonicalCandidateOperator(operator); candidate != "" {
			if candidate == "CANDIDATES" {
				plan.approximateCandidateCount++
			} else if plan.soleRootCandidateOperator == "" {
				plan.soleRootCandidateOperator = candidate
			}
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
		return nonNegated
	case isConditionType(node, "CompoundCondition"):
		conditions, ok := node["conditions"].([]any)
		if !ok {
			return false
		}
		operator, _ := node["operator"].(string)
		conjunction := nonNegated && strings.EqualFold(strings.TrimSpace(operator), "AND")
		for _, condition := range conditions {
			if !inspectConditionNode(condition, plan) {
				conjunction = false
			}
		}
		return conjunction
	default:
		return false
	}
}

func validateConditionPlan(raw json.RawMessage) error {
	plan, err := inspectConditionJSON(raw)
	if err != nil {
		return err
	}

	return validateInspectedConditionPlan(plan)
}

func validateInspectedConditionPlan(plan conditionPlan) error {
	if plan.soleRootCandidateOperator != "" &&
		(!plan.rootIsSingle ||
			plan.rootOperator != plan.soleRootCandidateOperator ||
			!plan.pureConjunction) {
		return fmt.Errorf("%s must be the sole root criterion", plan.soleRootCandidateOperator)
	}
	if plan.approximateCandidateCount > 1 {
		return fmt.Errorf("A query may contain only one CANDIDATES criterion")
	}
	if plan.approximateCandidateCount == 1 && !plan.pureConjunction {
		return fmt.Errorf("CANDIDATES can be combined only with non-negated AND predicates")
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

func combineConditionPlans(left, right conditionPlan, operator string) conditionPlan {
	return conditionPlan{
		soleRootCandidateOperator: firstNonEmpty(left.soleRootCandidateOperator, right.soleRootCandidateOperator),
		approximateCandidateCount: left.approximateCandidateCount + right.approximateCandidateCount,
		readOnlyOperator:          firstNonEmpty(left.readOnlyOperator, right.readOnlyOperator),
		searchCount:               left.searchCount + right.searchCount,
		searchWrongField:          left.searchWrongField || right.searchWrongField,
		fullTextCount:             left.fullTextCount + right.fullTextCount,
		pureConjunction: strings.EqualFold(strings.TrimSpace(operator), "AND") &&
			left.pureConjunction && right.pureConjunction,
	}
}

func firstNonEmpty(left, right string) string {
	if left != "" {
		return left
	}
	return right
}

func conditionNodeIsNegated(node map[string]any) bool {
	if trueConditionFlag(node, "isNot") || trueConditionFlag(node, "flip") {
		return true
	}
	criteria, ok := node["criteria"].(map[string]any)
	return ok && (trueConditionFlag(criteria, "isNot") || trueConditionFlag(criteria, "flip"))
}

func trueConditionFlag(node map[string]any, want string) bool {
	for key, value := range node {
		if strings.EqualFold(strings.TrimSpace(key), want) {
			flag, _ := value.(bool)
			return flag
		}
	}
	return false
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
