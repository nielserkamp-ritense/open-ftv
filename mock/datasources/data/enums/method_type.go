package enums

// MethodType represents an HTTP method.
type MethodType uint8

// List of supported HTTP methods.
const (
	GetMethod MethodType = iota + 1
	PostMethod
	PutMethod
	PatchMethod
	DeleteMethod
)

// String implements the Stringer interface.
func (t MethodType) String() string {
	switch t {
	case GetMethod:
		return "GET"
	case PostMethod:
		return "POST"
	case PutMethod:
		return "PUT"
	case PatchMethod:
		return "PATCH"
	case DeleteMethod:
		return "DELETE"
	default:
		return "[Unknown]"
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (t MethodType) MarshalJSON() ([]byte, error) {
	switch t {
	case GetMethod:
		return []byte(`"GET"`), nil
	case PostMethod:
		return []byte(`"POST"`), nil
	case PutMethod:
		return []byte(`"PUT"`), nil
	case PatchMethod:
		return []byte(`"PATCH"`), nil
	case DeleteMethod:
		return []byte(`"DELETE"`), nil
	default:
		return []byte(`"?invalid?"`), nil
	}
}

// MarshalYAML implements the YAML Marshaler interface.
func (t MethodType) MarshalYAML() ([]byte, error) {
	switch t {
	case GetMethod:
		return []byte("GET"), nil
	case PostMethod:
		return []byte("POST"), nil
	case PutMethod:
		return []byte("PUT"), nil
	case PatchMethod:
		return []byte("PATCH"), nil
	case DeleteMethod:
		return []byte("DELETE"), nil
	default:
		return []byte("?invalid?"), nil
	}
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (t *MethodType) UnmarshalJSON(b []byte) error {
	*t = MethodTypeFromString(string(b))
	return nil
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (t *MethodType) UnmarshalYAML(b []byte) error {
	*t = MethodTypeFromString(string(b))
	return nil
}

// MethodTypeFromString converts a string into an HTTP method.
func MethodTypeFromString(s string) MethodType {
	switch prepareFromString(s) {
	case "GET":
		return GetMethod
	case "POST":
		return PostMethod
	case "PUT":
		return PutMethod
	case "PATCH":
		return PatchMethod
	case "DELETE":
		return DeleteMethod
	default:
		return GetMethod
	}
}
