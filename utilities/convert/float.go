package convert

import (
	"strconv"
)

// AnyToFloat64 converts the input to a 64-bit floating point number.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it first converts the input to a string, before trying to parse it.
func AnyToFloat64(in any) float64 {
	var s string

	switch t := in.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case uint64:
		return float64(t)
	case bool:
		if t {
			return 1
		}
		return 0
	case string:
		s = t
	default:
		s = AnyToString(in)
	}

	i, _ := strconv.ParseFloat(s, 64)
	return i
}
