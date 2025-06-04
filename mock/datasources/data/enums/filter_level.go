package enums

import (
	"fmt"
	"strings"
)

// FilterLevel represents the level of a filter.
type FilterLevel uint8

// List of supported filter levels.
const (
	PrimaryLevel FilterLevel = iota + 1 // only on the primary table.
	AnyLevel                            // on the primary table or any join level.
	JoinLevel                           // only on a specific join level.
)

// String implements the Stringer interface.
func (l FilterLevel) String() string {
	switch l {
	case PrimaryLevel:
		return "primary"
	case AnyLevel:
		return "any"
	case JoinLevel:
		return "join"
	default:
		return "[unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (l FilterLevel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, l.String())), nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (l FilterLevel) MarshalYAML() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (l *FilterLevel) UnmarshalJSON(b []byte) error {
	*l = FilterLevelFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (l *FilterLevel) UnmarshalYAML(b []byte) error {
	*l = FilterLevelFromString(string(b))
	return nil
}

// FilterLevelFromString converts a string into a filter level.
func FilterLevelFromString(s string) FilterLevel {
	s = strings.NewReplacer(" ", "", "-", "", "_", "", "+", "", "&", "", "'", "", "\"", "", "`", "", "\n", "", "\t", "").Replace(s)

	switch strings.ToUpper(s) {
	case "ANY", "ANYLEVEL", "ALL", "ALLLEVEL", "ALLLEVELS":
		return AnyLevel
	case "JOIN", "JOINLEVEL":
		return JoinLevel
	default:
		return PrimaryLevel
	}
}
