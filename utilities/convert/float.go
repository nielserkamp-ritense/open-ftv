package convert

import (
	"fmt"
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

// AnyToFloats converts the input into a slice of floating point numbers.
func AnyToFloats(v any) []float64 {
	switch t := v.(type) {
	case float64:
		return []float64{t}
	case []float64:
		return t
	case nil:
		return []float64{}

	case []any:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []int:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []int8:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []int16:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []int32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []int64:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []uint:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []uint8:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []uint16:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []uint32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []uint64:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []float32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []bool:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	case []string:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = AnyToFloat64(t[i])
		}
		return out

	default:
		return []float64{AnyToFloat64(fmt.Sprintf("%v", v))}
	}
}
