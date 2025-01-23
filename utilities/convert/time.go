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

	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return t
	}
	if t, err := time.ParseInLocation("15:04:05.999999999", s, time.UTC); err == nil {
		return t
	}
	if t, err := time.ParseInLocation("2006-01", s, time.UTC); err == nil {
		return t
	}
	if t, err := time.ParseInLocation("01-02", s, time.UTC); err == nil {
		return t
	}
	return time.Time{}
}
