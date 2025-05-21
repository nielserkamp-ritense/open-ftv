package convert

import (
	"fmt"
	"time"
)

// AnyToDateTime converts the input to a time.Time variable.
//
// It supports the most commonly used Golang data-types directly.
// For other data-types it first converts the input to a string.
func AnyToDateTime(in any) time.Time {
	var s string

	switch t := in.(type) {
	case time.Time:
		return t
	case string:
		s = t
	default:
		s = fmt.Sprintf("%v", t)
	}

	for i := range formats {
		if t, err := time.Parse(formats[i], s); err == nil {
			return t
		}
	}

	return time.Time{}
}

var formats = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"2006-01-02",
	"2006/01/02",
	"20060102",
	"15:04:05.999999999",
	"150405.999999999",
	"150405.999",
	"150405",
	"2006-01",
	"2006/01",
	"200601",
	"01-02",
	"01/02",
	"0102",
}
