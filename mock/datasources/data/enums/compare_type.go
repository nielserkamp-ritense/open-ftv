package enums

import (
	"fmt"
)

// CompareType represents a type of comparison.
type CompareType uint8

// List of supported comparison types.
const (
	IsEqual CompareType = iota + 1
	IsNotEqual
	Exists
	NotExists
	IsLesser
	IsLesserOrEqual
	IsGreaterOrEqual
	IsGreater
	InList
	NotInList
	IsLike
	IsNotLike
	MatchRegex
	NotMatchRegex
	// hidden values
	cmpFirst = IsEqual
	cmpLast  = NotMatchRegex
)

// IsValid returns true if the comparison type is valid.
func (t CompareType) IsValid() bool {
	return t >= cmpFirst && t <= cmpLast
}

// String implements the Stringer interface.
func (t CompareType) String() string {
	switch t {
	case IsEqual:
		return "IsEqual"
	case IsNotEqual:
		return "IsNotEqual"
	case Exists:
		return "Exists"
	case NotExists:
		return "NotExists"
	case IsLesser:
		return "IsLesser"
	case IsLesserOrEqual:
		return "IsLesserOrEqual"
	case IsGreaterOrEqual:
		return "IsGreaterOrEqual"
	case IsGreater:
		return "IsGreater"
	case InList:
		return "InList"
	case NotInList:
		return "NotInList"
	case IsLike:
		return "IsLike"
	case IsNotLike:
		return "IsNotLike"
	case MatchRegex:
		return "MatchRegex"
	case NotMatchRegex:
		return "NotMatchRegex"
	default:
		return "[unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t CompareType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.String())), nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (t CompareType) MarshalYAML() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *CompareType) UnmarshalJSON(b []byte) error {
	*t = CompareTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *CompareType) UnmarshalYAML(b []byte) error {
	*t = CompareTypeFromString(string(b))
	return nil
}

// CompareTypeFromString converts a string into a comparison type.
func CompareTypeFromString(s string) CompareType {
	switch prepareFromString(s) {
	case "==", "=", "!NE", "EQ", "EQUAL", "ISEQUAL":
		return IsEqual
	case "!=", "<>", "NE", "!EQ", "NOTEQUAL", "ISNOTEQUAL":
		return IsNotEqual
	case "EXISTS", "!NIL", "NOTNIL", "ISNOTNIL", "NOTEMPTY", "ISNOTEMPTY":
		return Exists
	case "NOTEXISTS", "!EXISTS", "NIL", "ISNIL", "EMPTY", "ISEMPTY":
		return NotExists
	case "<", "!>=", "!=>", "LT", "!GE", "LESSER", "LESSERTHAN", "SMALLER", "SMALLERTHAN", "ISLESSER", "ISLESSERTHAN", "ISSMALLER", "ISSMALLERTHAN":
		return IsLesser
	case ">", "!<=", "!=<", "GT", "!LE", "GREATER", "GREATERTHAN", "ISGREATER", "ISGREATERTHAN":
		return IsGreater
	case "<=", "=<", "!>", "LE", "!GT", "LESSEROREQUAL", "LESSERTHANOREQUAL", "SMALLEROREQUAL", "SMALLERTHANOREQUAL", "ISLESSEROREQUAL", "ISLESSERTHANOREQUAL", "ISSMALLEROREQUAL", "ISSMALLERTHANOREQUAL":
		return IsLesserOrEqual
	case ">=", "=>", "!<", "GE", "!LT", "GREATEROREQUAL", "GREATERTHANOREQUAL", "ISGREATEROREQUAL", "ISGREATERTHANOREQUAL":
		return IsGreaterOrEqual
	case "IN", "INLIST":
		return InList
	case "!IN", "!INLIST", "NOTIN", "NOTINLIST":
		return NotInList
	case "LIKE", "ISLIKE", "WC", "WILDCARD":
		return IsLike
	case "!LIKE", "NOTLIKE", "ISNOTLIKE", "!WC", "!WILDCARD", "NOTWILDCARD":
		return IsNotLike
	case "~", "RX", "REGEX", "MATCHRX", "MATCHREGEX":
		return MatchRegex
	case "!~", "!RX", "!REGEX", "!MATCHRX", "!MATCHREGEX", "NOTRX", "NOTREGEX", "NOTMATCHRX", "NOTMATCHREGEX":
		return NotMatchRegex
	default:
		return 0
	}
}
