package enums

// OrderType represents a sort order.
type OrderType uint8

// list of supported sort orders.
const (
	OrderDescending OrderType = iota
	OrderAscending
)

// String implements the Stringer interface.
func (t OrderType) String() string {
	switch t {
	case OrderAscending:
		return "Ascending"
	default:
		return "Descending"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t OrderType) MarshalJSON() ([]byte, error) {
	switch t {
	case OrderAscending:
		return []byte(`"ASC"`), nil
	default:
		return []byte(`"DESC"`), nil
	}
}

// MarshalYAML implements the YAML Marshaler interface.
func (t OrderType) MarshalYAML() ([]byte, error) {
	switch t {
	case OrderAscending:
		return []byte("ASC"), nil
	default:
		return []byte("DESC"), nil
	}
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *OrderType) UnmarshalJSON(b []byte) error {
	*t = OrderTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *OrderType) UnmarshalYAML(b []byte) error {
	return t.UnmarshalJSON(b)
}

// OrderTypeFromString converts a string into a sort order.
func OrderTypeFromString(s string) OrderType {
	switch prepareFromString(s) {
	case "ASC", "ASCENDING":
		return OrderAscending
	default:
		return OrderDescending
	}
}
