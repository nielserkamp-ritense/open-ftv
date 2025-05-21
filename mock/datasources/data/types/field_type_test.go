package types

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want FieldType
	}{
		{name: "x", want: StringType},
		{name: "bad name", want: StringType},
		{name: "string", want: StringType},
		{name: "varchar", want: StringType},
		{name: "int", want: IntegerType},
		{name: "uint8", want: UnsignedIntegerType},
		{name: "unsigned integer", want: UnsignedIntegerType},
		{name: "int16", want: IntegerType},
		{name: "float", want: FloatType},
		{name: "float_64", want: FloatType},
		{name: "double", want: FloatType},
		{name: "bool", want: BooleanType},
		{name: "boolean", want: BooleanType},
		{name: "date", want: DateType},
		{name: "time", want: TimeType},
		{name: "date time", want: DateTimeType},
		{name: "date + time", want: DateTimeType},
		{name: "url", want: URLType},
		{name: "email", want: EmailType},
		{name: "phone", want: PhoneNrType},
		{name: "ip address", want: IPAddressType},
		{name: "ip-addr", want: IPAddressType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FieldTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFieldType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FieldType
		want string
	}{
		{name: "unknown", t: 199, want: "[Unknown]"},
		{name: "string", t: StringType, want: "String"},
		{name: "integer", t: IntegerType, want: "Integer"},
		{name: "unsigned", t: UnsignedIntegerType, want: "Unsigned integer"},
		{name: "bool", t: BooleanType, want: "Boolean"},
		{name: "float", t: FloatType, want: "Float"},
		{name: "date", t: DateType, want: "Date"},
		{name: "time", t: TimeType, want: "Time"},
		{name: "date+time", t: DateTimeType, want: "Date and time"},
		{name: "url", t: URLType, want: "URL"},
		{name: "email", t: EmailType, want: "Email address"},
		{name: "phone", t: PhoneNrType, want: "Phone number"},
		{name: "ip", t: IPAddressType, want: "IP address"},
		{name: "object", t: ObjectType, want: "Object"},
		{name: "any", t: AnyType, want: "Any"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFieldType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FieldType
	}{
		{name: "string", t: StringType},
		{name: "integer", t: IntegerType},
		{name: "unsigned", t: UnsignedIntegerType},
		{name: "bool", t: BooleanType},
		{name: "float", t: FloatType},
		{name: "date", t: DateType},
		{name: "time", t: TimeType},
		{name: "date+time", t: DateTimeType},
		{name: "url", t: URLType},
		{name: "email", t: EmailType},
		{name: "phone", t: PhoneNrType},
		{name: "ip", t: IPAddressType},
		{name: "any", t: AnyType},
		{name: "object", t: ObjectType},
		{name: "class", t: ObjectType},
		{name: "struct", t: ObjectType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 FieldType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestFieldType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FieldType
	}{
		{name: "string", t: StringType},
		{name: "integer", t: IntegerType},
		{name: "unsigned", t: UnsignedIntegerType},
		{name: "bool", t: BooleanType},
		{name: "float", t: FloatType},
		{name: "date", t: DateType},
		{name: "time", t: TimeType},
		{name: "date+time", t: DateTimeType},
		{name: "url", t: URLType},
		{name: "email", t: EmailType},
		{name: "phone", t: PhoneNrType},
		{name: "ip", t: IPAddressType},
		{name: "any", t: AnyType},
		{name: "object", t: ObjectType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 FieldType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestFieldType_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		f := FieldType(199)

		got, err := f.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, `"?invalid?"`, string(got))

		got, err = f.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, "?invalid?", string(got))
	})
}
