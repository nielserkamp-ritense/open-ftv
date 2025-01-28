package convert

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

// OpaqueString returns the actual string from the given string pointer,
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
