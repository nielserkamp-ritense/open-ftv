package enums

// TransformationType represents a transformation type.
type TransformationType uint8

// List of supported transformation types.
const (
	TransformCompare TransformationType = iota + 1
	TransformConvert
	TransformAge
	// future extensions
	// TransformAdd
	// TransformSubtract
	// TransformMultiply
	// TransformDivide
	// TransformRound
	// TransformAnd
	// TransformOr
	// TransformNot
	// TransformUpper
	// TransformLower
	// TransformBase64
	// TransformBinHex
	// TransformMask
	// TransformHash
	// TransformEncrypt
)

// String implements the Stringer interface.
func (t TransformationType) String() string {
	switch t {
	case TransformCompare:
		return "Compare"
	case TransformConvert:
		return "Convert"
	case TransformAge:
		return "Age"
	default:
		return "[Unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t TransformationType) MarshalJSON() ([]byte, error) {
	switch t {
	case TransformCompare:
		return []byte(`"Compare"`), nil
	case TransformConvert:
		return []byte(`"Convert"`), nil
	case TransformAge:
		return []byte(`"Age"`), nil
	default:
		return []byte(`"?invalid?"`), nil
	}
}

// MarshalYAML implements the YAML Marshaler interface.
func (t TransformationType) MarshalYAML() ([]byte, error) {
	switch t {
	case TransformCompare:
		return []byte("Compare"), nil
	case TransformConvert:
		return []byte("Convert"), nil
	case TransformAge:
		return []byte("Age"), nil
	default:
		return []byte("?invalid?"), nil
	}
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *TransformationType) UnmarshalJSON(b []byte) error {
	*t = TransformationTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *TransformationType) UnmarshalYAML(b []byte) error {
	*t = TransformationTypeFromString(string(b))
	return nil
}

// TransformationTypeFromString converts a string into an HTTP method.
func TransformationTypeFromString(s string) TransformationType {
	switch prepareFromString(s) {
	case "COMPARE":
		return TransformCompare
	case "CONVERT":
		return TransformConvert
	case "AGE":
		return TransformAge
	default:
		return TransformCompare
	}
}
