// Package convert contains conversion routines.
package convert

import (
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
