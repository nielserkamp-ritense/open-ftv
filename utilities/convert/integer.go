package convert

import (
	"math"
	"strconv"
)

// AnyToInt64 converts the input to a 64-bit signed integer.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it first converts the input to a string, before trying to parse it.
func AnyToInt64(in any) int64 {
	var s string

	switch t := in.(type) {
	case int:
		return int64(t)
	case int64:
		return t
	case uint:
		return int64(t)
	case uint64:
		return int64(t)
	case float64:
		return int64(math.Round(t))
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

	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// AnyToUint64 converts the input to a 64-bit unsigned integer.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it first converts the input to a string, before trying to parse it.
func AnyToUint64(in any) uint64 {
	var s string

	switch t := in.(type) {
	case uint:
		return uint64(t)
	case uint64:
		return t
	case int:
		return uint64(t)
	case int64:
		return uint64(t)
	case float64:
		return uint64(math.Round(t))
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

	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}
