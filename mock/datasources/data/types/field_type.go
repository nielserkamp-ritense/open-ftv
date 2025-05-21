// Package types contains type definitions.
package types

import "strings"

// FieldType represents a field type.
type FieldType uint8

// List of field types.
const (
	StringType FieldType = iota
	IntegerType
	UnsignedIntegerType
	FloatType
	BooleanType
	DateType
	TimeType
	DateTimeType
	URLType
	EmailType
	PhoneNrType
	IPAddressType
	AnyType    = 98
	ObjectType = 99
)

// String implements the Stringer interface.
func (t FieldType) String() string {
	switch t {
	case StringType:
		return "String"
	case IntegerType:
		return "Integer"
	case UnsignedIntegerType:
		return "Unsigned integer"
	case FloatType:
		return "Float"
	case BooleanType:
		return "Boolean"
	case DateType:
		return "Date"
	case TimeType:
		return "Time"
	case DateTimeType:
		return "Date and time"
	case URLType:
		return "URL"
	case EmailType:
		return "Email address"
	case PhoneNrType:
		return "Phone number"
	case IPAddressType:
		return "IP address"
	case AnyType:
		return "Any"
	case ObjectType:
		return "Object"
	default:
		return "[Unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t FieldType) MarshalJSON() ([]byte, error) {
	switch t {
	case StringType:
		return []byte(`"string"`), nil
	case IntegerType:
		return []byte(`"int"`), nil
	case UnsignedIntegerType:
		return []byte(`"uint"`), nil
	case FloatType:
		return []byte(`"float"`), nil
	case BooleanType:
		return []byte(`"bool"`), nil
	case DateType:
		return []byte(`"date"`), nil
	case TimeType:
		return []byte(`"time"`), nil
	case DateTimeType:
		return []byte(`"datetime"`), nil
	case URLType:
		return []byte(`"url"`), nil
	case EmailType:
		return []byte(`"email"`), nil
	case PhoneNrType:
		return []byte(`"phone"`), nil
	case IPAddressType:
		return []byte(`"ipaddr"`), nil
	case AnyType:
		return []byte(`"any"`), nil
	case ObjectType:
		return []byte(`"object"`), nil
	default:
		return []byte(`"?invalid?"`), nil
	}
}

// MarshalYAML implements the YAML Marshaler interface.
func (t FieldType) MarshalYAML() ([]byte, error) {
	switch t {
	case StringType:
		return []byte("string"), nil
	case IntegerType:
		return []byte("int"), nil
	case UnsignedIntegerType:
		return []byte("uint"), nil
	case FloatType:
		return []byte("float"), nil
	case BooleanType:
		return []byte("bool"), nil
	case DateType:
		return []byte("date"), nil
	case TimeType:
		return []byte("time"), nil
	case DateTimeType:
		return []byte("datetime"), nil
	case URLType:
		return []byte("url"), nil
	case EmailType:
		return []byte("email"), nil
	case PhoneNrType:
		return []byte("phone"), nil
	case IPAddressType:
		return []byte("ipaddr"), nil
	case AnyType:
		return []byte("any"), nil
	case ObjectType:
		return []byte("object"), nil
	default:
		return []byte("?invalid?"), nil
	}
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *FieldType) UnmarshalJSON(b []byte) error {
	*t = FieldTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *FieldType) UnmarshalYAML(b []byte) error {
	*t = FieldTypeFromString(string(b))
	return nil
}

// FieldTypeFromString converts a string into a sort order.
func FieldTypeFromString(s string) FieldType {
	s = strings.NewReplacer(" ", "", "-", "", "_", "", "+", "", "&", "", "'", "", "\"", "", "`", "").Replace(s)

	switch strings.ToUpper(s) {
	case "INTEGER", "INT", "INT8", "INT16", "INT32", "INT64":
		return IntegerType
	case "UNSIGNEDINTEGER", "UNSIGNED", "UINT", "UINT8", "UINT16", "UINT32", "UINT64":
		return UnsignedIntegerType
	case "FLOAT", "DOUBLE", "FLOAT32", "FLOAT64":
		return FloatType
	case "BOOLEAN", "BOOL":
		return BooleanType
	case "DATE":
		return DateType
	case "TIME":
		return TimeType
	case "DATETIME", "TIMESTAMP":
		return DateTimeType
	case "URL":
		return URLType
	case "EMAIL":
		return EmailType
	case "PHONE":
		return PhoneNrType
	case "IPADDRESS", "IPADDR":
		return IPAddressType
	case "ANY":
		return AnyType
	case "OBJECT", "OBJ", "CLASS", "RECORD", "STRUCT":
		return ObjectType
	default:
		return StringType
	}
}
