// Package meta — label and field selector parsing and matching.
package meta

import (
	"fmt"
	"strings"
)

// ─── LabelSelector ───────────────────────────────────────────────────────────

// LabelSelector is a label query over a set of resources.
type LabelSelector struct {
	// MatchLabels is a map of {key,value} pairs.  All requirements in the map
	// must be satisfied for the object to match.
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	// MatchExpressions is a list of label selector requirements.
	MatchExpressions []LabelSelectorRequirement `json:"matchExpressions,omitempty"`
}

// LabelSelectorRequirement is a selector that contains values, a key, and an
// operator that relates the key and values.
type LabelSelectorRequirement struct {
	// Key is the label key that the selector applies to.
	Key string `json:"key"`
	// Operator represents a key's relationship to a set of values.
	Operator LabelSelectorOperator `json:"operator"`
	// Values is an array of string values.
	Values []string `json:"values,omitempty"`
}

// LabelSelectorOperator is a set of supported operators.
type LabelSelectorOperator string

const (
	LabelSelectorOpIn           LabelSelectorOperator = "In"
	LabelSelectorOpNotIn        LabelSelectorOperator = "NotIn"
	LabelSelectorOpExists       LabelSelectorOperator = "Exists"
	LabelSelectorOpDoesNotExist LabelSelectorOperator = "DoesNotExist"
)

// ─── ParsedLabelSelector ─────────────────────────────────────────────────────

// ParsedLabelSelector is a compiled label selector that can match object labels.
type ParsedLabelSelector struct {
	requirements []labelRequirement
}

type labelRequirement struct {
	key      string
	op       LabelSelectorOperator
	values   map[string]struct{}
}

// ParseLabelSelector parses a string representation of a label selector into
// a ParsedLabelSelector.  The syntax is a comma-separated list of requirements.
//
// Supported forms:
//   - key=value      (equality)
//   - key==value     (equality, kubectl shorthand)
//   - key!=value     (inequality)
//   - key            (existence)
//   - !key           (non-existence)
//   - key in (v1,v2) (set-based)
//   - key notin (v1,v2)
func ParseLabelSelector(selector string) (*ParsedLabelSelector, error) {
	if selector == "" {
		return &ParsedLabelSelector{}, nil
	}
	var reqs []labelRequirement
	for _, part := range splitCommaRespectingParens(selector) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		req, err := parseLabelRequirement(part)
		if err != nil {
			return nil, fmt.Errorf("invalid label selector %q: %w", selector, err)
		}
		reqs = append(reqs, req)
	}
	return &ParsedLabelSelector{requirements: reqs}, nil
}

// Matches reports whether the given labels satisfy the selector.
func (s *ParsedLabelSelector) Matches(labels map[string]string) bool {
	for _, req := range s.requirements {
		if !req.matches(labels) {
			return false
		}
	}
	return true
}

// Empty reports whether the selector is effectively empty (matches everything).
func (s *ParsedLabelSelector) Empty() bool {
	return len(s.requirements) == 0
}

func (r labelRequirement) matches(labels map[string]string) bool {
	val, exists := labels[r.key]
	switch r.op {
	case LabelSelectorOpIn:
		if !exists {
			return false
		}
		_, ok := r.values[val]
		return ok
	case LabelSelectorOpNotIn:
		if !exists {
			return true
		}
		_, ok := r.values[val]
		return !ok
	case LabelSelectorOpExists:
		return exists
	case LabelSelectorOpDoesNotExist:
		return !exists
	}
	return false
}

func parseLabelRequirement(s string) (labelRequirement, error) {
	// !key  — DoesNotExist
	if strings.HasPrefix(s, "!") {
		key := strings.TrimPrefix(s, "!")
		if err := validateLabelKey(key); err != nil {
			return labelRequirement{}, err
		}
		return labelRequirement{key: key, op: LabelSelectorOpDoesNotExist}, nil
	}

	// key in (v1,v2)
	if i := strings.Index(strings.ToLower(s), " in ("); i >= 0 {
		key := strings.TrimSpace(s[:i])
		rest := s[i+len(" in ("):]
		rest = strings.TrimSuffix(rest, ")")
		vals, err := parseValues(rest)
		if err != nil {
			return labelRequirement{}, err
		}
		return labelRequirement{key: key, op: LabelSelectorOpIn, values: vals}, nil
	}

	// key notin (v1,v2)
	if i := strings.Index(strings.ToLower(s), " notin ("); i >= 0 {
		key := strings.TrimSpace(s[:i])
		rest := s[i+len(" notin ("):]
		rest = strings.TrimSuffix(rest, ")")
		vals, err := parseValues(rest)
		if err != nil {
			return labelRequirement{}, err
		}
		return labelRequirement{key: key, op: LabelSelectorOpNotIn, values: vals}, nil
	}

	// key!=value
	if i := strings.Index(s, "!="); i >= 0 {
		key := strings.TrimSpace(s[:i])
		val := strings.TrimSpace(s[i+2:])
		return labelRequirement{key: key, op: LabelSelectorOpNotIn, values: map[string]struct{}{val: {}}}, nil
	}

	// key==value or key=value
	if i := strings.IndexAny(s, "="); i >= 0 {
		key := strings.TrimSpace(strings.TrimSuffix(s[:i], "="))
		val := strings.TrimSpace(s[i+1:])
		if strings.HasPrefix(val, "=") {
			val = val[1:]
		}
		return labelRequirement{key: key, op: LabelSelectorOpIn, values: map[string]struct{}{val: {}}}, nil
	}

	// bare key — Exists
	if err := validateLabelKey(s); err != nil {
		return labelRequirement{}, err
	}
	return labelRequirement{key: s, op: LabelSelectorOpExists}, nil
}

func parseValues(s string) (map[string]struct{}, error) {
	m := make(map[string]struct{})
	for _, v := range strings.Split(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return m, nil
}

func validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty label key")
	}
	return nil
}

// splitCommaRespectingParens splits s by commas, but does not split inside parentheses.
func splitCommaRespectingParens(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

// ─── FieldSelector ───────────────────────────────────────────────────────────

// ParsedFieldSelector is a compiled field selector that can match object fields.
// In Phase 1, only metadata.name and metadata.namespace are supported as indexed fields.
// Arbitrary field selectors require a full scan.
type ParsedFieldSelector struct {
	requirements []fieldRequirement
}

type fieldRequirement struct {
	field string
	op    string // "=" or "!="
	value string
}

// ParseFieldSelector parses a field selector string.
//
// Supported forms:
//   - field=value
//   - field==value
//   - field!=value
func ParseFieldSelector(selector string) (*ParsedFieldSelector, error) {
	if selector == "" {
		return &ParsedFieldSelector{}, nil
	}
	var reqs []fieldRequirement
	for _, part := range strings.Split(selector, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		req, err := parseFieldRequirement(part)
		if err != nil {
			return nil, fmt.Errorf("invalid field selector %q: %w", selector, err)
		}
		reqs = append(reqs, req)
	}
	return &ParsedFieldSelector{requirements: reqs}, nil
}

// Matches reports whether the given fields (extracted from an object) match the selector.
func (s *ParsedFieldSelector) Matches(fields map[string]string) bool {
	for _, req := range s.requirements {
		val := fields[req.field]
		switch req.op {
		case "=", "==":
			if val != req.value {
				return false
			}
		case "!=":
			if val == req.value {
				return false
			}
		}
	}
	return true
}

// Empty reports whether the selector is effectively empty.
func (s *ParsedFieldSelector) Empty() bool {
	return len(s.requirements) == 0
}

// Requirements returns the field requirements.
func (s *ParsedFieldSelector) Requirements() []struct{ Field, Op, Value string } {
	out := make([]struct{ Field, Op, Value string }, len(s.requirements))
	for i, r := range s.requirements {
		out[i] = struct{ Field, Op, Value string }{r.field, r.op, r.value}
	}
	return out
}

func parseFieldRequirement(s string) (fieldRequirement, error) {
	// != must be checked before = to avoid partial match.
	if i := strings.Index(s, "!="); i >= 0 {
		field := strings.TrimSpace(s[:i])
		value := strings.TrimSpace(s[i+2:])
		return fieldRequirement{field: field, op: "!=", value: value}, nil
	}
	if i := strings.Index(s, "=="); i >= 0 {
		field := strings.TrimSpace(s[:i])
		value := strings.TrimSpace(s[i+2:])
		return fieldRequirement{field: field, op: "==", value: value}, nil
	}
	if i := strings.Index(s, "="); i >= 0 {
		field := strings.TrimSpace(s[:i])
		value := strings.TrimSpace(s[i+1:])
		return fieldRequirement{field: field, op: "=", value: value}, nil
	}
	return fieldRequirement{}, fmt.Errorf("unsupported field selector term %q: expected field=value, field==value, or field!=value", s)
}

// ExtractObjectFields extracts the indexed fields from an ObjectMeta for field selector matching.
func ExtractObjectFields(meta *ObjectMeta) map[string]string {
	fields := make(map[string]string, 4)
	fields["metadata.name"] = meta.Name
	fields["metadata.namespace"] = meta.namespace()
	return fields
}

// namespace returns the namespace or "default" if empty.
func (o *ObjectMeta) namespace() string {
	if o.Namespace == "" {
		return ""
	}
	return o.Namespace
}
