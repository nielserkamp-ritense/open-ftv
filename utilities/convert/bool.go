// Package convert contains conversion routines.
package convert

import (
	"fmt"
	"strings"
)

// AnyToBool converts the input to a boolean.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it first converts the input to a string.
//
// Any numeric value other than zero, is considered true.
// A string with value "true" or "1" is considered true.
// All other values for a string are considered false.
func AnyToBool(in any) bool {
	var s string

	switch t := in.(type) {
	case bool:
		return t
	case string:
		s = t
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0.0
	default:
		s = AnyToString(in)
	}

	return strings.EqualFold(s, "true") || s == "1"
}

// AnyToBools converts the input into a slice of booleans.
func AnyToBools(v any) []bool {
	switch t := v.(type) {
	case bool:
		return []bool{t}
	case []bool:
		return t
	case nil:
		return []bool{}

	case []any:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []int:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []int8:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []int16:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []int32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []int64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []uint:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []uint8:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []uint16:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []uint32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []uint64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []float32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []float64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	case []string:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = AnyToBool(t[i])
		}
		return out

	default:
		return []bool{AnyToBool(fmt.Sprintf("%v", v))}
	}
}
