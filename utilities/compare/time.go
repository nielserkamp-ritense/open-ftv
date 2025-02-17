// Package compare contains comparison functions.
package compare

import "time"

// TimePtrEqual returns true if both timestamps are nil or both contain the exact same value.
func TimePtrEqual(d1, d2 *time.Time) bool {
	switch {
	case d1 == nil:
		return d2 == nil
	case d2 == nil:
		return false
	default:
		return d1.Equal(*d2)
	}
}
