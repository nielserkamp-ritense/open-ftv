package csv

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// anyToBools converts the input into a slice of booleans.
func anyToBools(v any) []bool {
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
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []int:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []int8:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []int16:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []int32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []int64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []uint:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []uint8:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []uint16:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []uint32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []uint64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []float32:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []float64:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	case []string:
		out := make([]bool, len(t))
		for i := range t {
			out[i] = convert.AnyToBool(t[i])
		}
		return out

	default:
		return []bool{convert.AnyToBool(fmt.Sprintf("%v", v))}
	}
}

// anyToFloats converts the input into a slice of floating point numbers.
func anyToFloats(v any) []float64 {
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
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []int:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []int8:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []int16:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []int32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []int64:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []uint:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []uint8:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []uint16:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []uint32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []uint64:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []float32:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []bool:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	case []string:
		out := make([]float64, len(t))
		for i := range t {
			out[i] = convert.AnyToFloat64(t[i])
		}
		return out

	default:
		return []float64{convert.AnyToFloat64(fmt.Sprintf("%v", v))}
	}
}

// anyToInts converts the input into a slice of integers.
func anyToInts(v any) []int64 {
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
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []int:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []int8:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []int16:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []int32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []uint:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []uint8:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []uint16:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []uint32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []uint64:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []float32:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []float64:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []bool:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	case []string:
		out := make([]int64, len(t))
		for i := range t {
			out[i] = convert.AnyToInt64(t[i])
		}
		return out

	default:
		return []int64{convert.AnyToInt64(fmt.Sprintf("%v", v))}
	}
}

// anyToUints converts the input into a slice of unsigned integers.
func anyToUints(v any) []uint64 {
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
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []int:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []int8:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []int16:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []int32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []int64:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []uint:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []uint8:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []uint16:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []uint32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []float32:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []float64:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []bool:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	case []string:
		out := make([]uint64, len(t))
		for i := range t {
			out[i] = convert.AnyToUint64(t[i])
		}
		return out

	default:
		return []uint64{convert.AnyToUint64(fmt.Sprintf("%v", v))}
	}
}
