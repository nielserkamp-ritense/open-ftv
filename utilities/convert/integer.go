package convert

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// PrependZero10 returns the input number as a decimal encoded string.
//
// If the length of n, expressed as a decimal number, is less than size,
// the output is prepended with zeroes until it is of size length.
func PrependZero10(n, size uint) string {
	return fmt.Sprintf("%0*d", size, n)
}

// MustIntDecimal converts the input string to an integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an integer, zero is returned.
func MustIntDecimal(input string) int {
	if i, err := FromIntDecimal(input); err == nil {
		return i
	}
	return 0
}

// FromIntDecimal converts the input string to an integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an integer, the parser error is returned.
func FromIntDecimal(input string) (int, error) {
	i, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

// MustUintDecimal converts the input string to an unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer, zero is returned.
func MustUintDecimal(input string) uint {
	if i, err := FromUintDecimal(input); err == nil {
		return i
	}
	return 0
}

// FromUintDecimal converts the input string to an unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer, the parser error is returned.
func FromUintDecimal(input string) (uint, error) {
	i, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

// MustUint8Decimal converts the input string to an 8-bit unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer or does not fit in 8 bits, zero is returned.
func MustUint8Decimal(input string) uint8 {
	if i, err := FromUintDecimal(input); err == nil && i <= math.MaxUint8 {
		return uint8(i)
	}
	return 0
}

// FromUint8Decimal converts the input string to an 8-bit unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer, the parser error is returned.
// If the numeric value does not fit in 8 bits, an error is returned.
func FromUint8Decimal(input string) (uint8, error) {
	i, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, err
	}
	if i > math.MaxUint8 {
		return 0, errors.New("value out of bounds")
	}
	return uint8(i), nil
}

// MustUint32Decimal converts the input string to a 32-bit unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer or does not fit in 32 bits, zero is returned.
func MustUint32Decimal(input string) uint32 {
	if i, err := FromUintDecimal(input); err == nil && i <= math.MaxUint32 {
		return uint32(i)
	}
	return 0
}

// FromUint32Decimal converts the input string to a 32-bit unsigned integer using decimal decoding.
//
// Any white space in the input is discarded before parsing.
// If the string starts with one or more zeroes, they are discarded.
// If the string cannot be parsed to an unsigned integer, the parser error is returned.
// If the numeric value does not fit in 32 bits, an error is returned.
func FromUint32Decimal(input string) (uint32, error) {
	i, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, err
	}
	if i > math.MaxUint32 {
		return 0, errors.New("value out of bounds")
	}
	return uint32(i), nil
}

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
