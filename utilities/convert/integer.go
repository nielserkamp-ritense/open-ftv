package convert

import (
	"fmt"
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

// AnyToInts converts the input into a slice of integers.
func AnyToInts(v any) []int64 {
	switch t := v.(type) {
	case int64:
		return []int64{t}
	case []int64:
		return t
	case nil:
		return []int64{}

	case []any:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []int:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []int8:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []int16:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []int32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []uint:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []uint8:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []uint16:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []uint32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []uint64:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []float32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []float64:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []bool:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	case []string:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = AnyToInt64(t[i])
		}
		return out

	default:
		return []int64{AnyToInt64(fmt.Sprintf("%v", v))}
	}
}

// AnyToUints converts the input into a slice of unsigned integers.
func AnyToUints(v any) []uint64 {
	switch t := v.(type) {
	case uint64:
		return []uint64{t}
	case []uint64:
		return t
	case nil:
		return []uint64{}

	case []any:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []int:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []int8:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []int16:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []int32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []int64:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []uint:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []uint8:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []uint16:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []uint32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []float32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []float64:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []bool:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	case []string:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = AnyToUint64(t[i])
		}
		return out

	default:
		return []uint64{AnyToUint64(fmt.Sprintf("%v", v))}
	}
}
