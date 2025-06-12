package enums

import (
	"fmt"
)

// JoinType represents a table join type.
type JoinType uint8

// List of supported table join types.
const (
	OptionalParentChild JoinType = iota + 1 // represents a parent/child relationship; joined table will be added as an array.
	OptionalSibling                         // represents a sibling relationship; joined table will be merged, and all fields will be qualified.
	ForcedParentChild                       // represents a forced parent/child relationship; joined table will be added as an array, even if no matched records are found.
	ForcedSibling                           // represents a forced sibling relationship; joined table will be merged, even if it is not found.
)

// String implements the Stringer interface.
func (t JoinType) String() string {
	switch t {
	case OptionalParentChild:
		return "OptionalParentChild"
	case OptionalSibling:
		return "OptionalSibling"
	case ForcedParentChild:
		return "ForcedParentChild"
	case ForcedSibling:
		return "ForcedSibling"
	default:
		return "[unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t JoinType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.String())), nil
}

// MarshalYAML implements the YAML Marshaler interface.
func (t JoinType) MarshalYAML() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *JoinType) UnmarshalJSON(b []byte) error {
	*t = JoinTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *JoinType) UnmarshalYAML(b []byte) error {
	*t = JoinTypeFromString(string(b))
	return nil
}

// JoinTypeFromString converts a string into a table join type.
func JoinTypeFromString(s string) JoinType {
	switch prepareFromString(s) {
	case "OPTIONALSIBLING", "SIBLING":
		return OptionalSibling
	case "FORCEDPARENTCHILD", "FORCED", "FORCEDPARENT", "FORCEDCHILD":
		return ForcedParentChild
	case "FORCEDSIBLING":
		return ForcedSibling
	default:
		return OptionalParentChild
	}
}
