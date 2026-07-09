package odrl_geo

import (
	"encoding/json"
	"strconv"
	"strings"
)

// triState is the outcome of evaluating a single constraint: satisfied, not
// satisfied, or not evaluable (spec.md §7.5, fail-safe).
type triState uint8

const (
	tUnknown triState = iota // not evaluable
	tTrue
	tFalse
)

func boolTri(b bool) triState {
	if b {
		return tTrue
	}
	return tFalse
}

// compareProperty evaluates a feature-property operator against the actual
// value and the constraint literals. It returns tUnknown when the value cannot
// be compared (e.g. non-numeric value for an ordering operator).
func compareProperty(op string, actual any, want []literal) triState {
	switch op {
	case opEq:
		return boolTri(valueEquals(actual, want))
	case opNeq:
		return boolTri(!valueEquals(actual, want))
	case opIsAnyOf:
		return boolTri(valueEquals(actual, want)) // any-of == equals against the list.
	case opIsAllOf:
		return boolTri(valueContainsAll(actual, want))
	case opLt, opLteq, opGt, opGteq:
		return compareOrder(op, actual, want)
	}
	return tUnknown
}

// valueEquals reports whether actual equals any of the wanted literals
// (numeric-aware, else string-wise).
func valueEquals(actual any, want []literal) bool {
	for _, w := range want {
		if af, aok := toFloat(actual); aok {
			if wf, wok := toFloat(w.Value); wok && af == wf {
				return true
			}
		}
		if toString(actual) == w.Value {
			return true
		}
	}
	return false
}

// valueContainsAll reports whether the actual (a set/list or string) contains
// every wanted literal (used for odrl:isAllOf).
func valueContainsAll(actual any, want []literal) bool {
	set := map[string]bool{}
	switch v := actual.(type) {
	case []any:
		for _, e := range v {
			set[toString(e)] = true
		}
	case []string:
		for _, e := range v {
			set[e] = true
		}
	default:
		for _, e := range strings.Split(toString(actual), ",") {
			set[strings.TrimSpace(e)] = true
		}
	}
	for _, w := range want {
		if !set[w.Value] {
			return false
		}
	}
	return len(want) > 0
}

// compareOrder evaluates an ordering operator numerically.
func compareOrder(op string, actual any, want []literal) triState {
	if len(want) == 0 {
		return tUnknown
	}
	af, aok := toFloat(actual)
	wf, wok := toFloat(want[0].Value)
	if !aok || !wok {
		return tUnknown
	}
	switch op {
	case opLt:
		return boolTri(af < wf)
	case opLteq:
		return boolTri(af <= wf)
	case opGt:
		return boolTri(af > wf)
	case opGteq:
		return boolTri(af >= wf)
	}
	return tUnknown
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f, err == nil
	}
	return 0, false
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	case json.Number:
		return x.String()
	}
	return ""
}

// coerceValue converts an obligation refinement value to a number when it looks
// numeric, so obligations carry typed properties (spec.md §7.3 example).
func coerceValue(s string) any {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}
