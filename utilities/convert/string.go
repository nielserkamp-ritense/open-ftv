package convert

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

// OpaqueString returns the actual string from the given string pointer
// or an empty string if the input is nil.
func OpaqueString(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

// RemoveHeaderParameters removes any optional parameters from an HTTP header value.
func RemoveHeaderParameters(in string) string {
	if i := strings.Index(in, ";"); i >= 0 {
		return strings.TrimSpace(in[:i])
	}
	return in
}

// AnyToString converts the input to a string.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it uses fmt.Sprintf("%v", in).
func AnyToString(in any) string {
	switch t := in.(type) {
	case string:
		return t
	case nil:
		return ""
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", in)
	}
}

// ForceSuffix returns the input with an appended suffix, regardless of whether the input contained the suffix or not.
func ForceSuffix(in, suffix string) string {
	if !strings.HasSuffix(in, suffix) {
		return in + suffix
	}
	return in
}

// AnyToStrings converts the input into a slice of strings.
func AnyToStrings(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	case nil:
		return []string{}

	case []any:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []int:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []int8:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []int16:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []int32:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []int64:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []uint:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []uint8:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []uint16:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []uint32:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []uint64:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []float32:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []float64:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	case []bool:
		out := make([]string, len(t))
		for i := range t {
			out[i] = AnyToString(t[i])
		}
		return out

	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}
