package compare

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestLesser(t *testing.T) {
	t.Parallel()

	var s1 = struct {
		Name string
		Num  int
	}{Name: "one", Num: 1}

	var s2 = struct {
		Name string
		Num  int
	}{Name: "One", Num: 1}

	var s3 = struct {
		Name string
		Num  int
	}{Name: "one", Num: 1}

	testcases := []struct {
		name        string
		t           enums.FieldType
		insensitive bool
		v1          any
		v2          any
		want        bool
	}{
		{name: "string, both nil", t: enums.StringType},
		{name: "string, first nil", t: enums.StringType, v2: "haha", want: true},
		{name: "string, second nil", t: enums.StringType, v1: "haha"},
		{name: "string, not lesser", t: enums.StringType, v1: "hihi", v2: "Hihi"},
		{name: "string, lesser", t: enums.StringType, v1: "Hoho", v2: "hoho", want: true},
		{name: "string, equalFold", t: enums.StringType, insensitive: true, v1: "hOhO", v2: "HoHo"},
		{name: "integer, both nil", t: enums.IntegerType},
		{name: "integer, first nil", t: enums.IntegerType, v2: "-1"},
		{name: "integer, second nil", t: enums.IntegerType, v1: "0"},
		{name: "integer, lesser", t: enums.IntegerType, v1: "123", v2: uint8(124), want: true},
		{name: "integer, not lesser", t: enums.IntegerType, v1: "-123", v2: -123.0},
		{name: "unsigned, both nil", t: enums.UnsignedIntegerType},
		{name: "unsigned, first nil", t: enums.UnsignedIntegerType, v2: "1", want: true},
		{name: "unsigned, second nil", t: enums.UnsignedIntegerType, v1: "0"},
		{name: "unsigned, not lesser", t: enums.UnsignedIntegerType, v1: "123", v2: 122.4999},
		{name: "unsigned, lesser", t: enums.UnsignedIntegerType, v1: "123", v2: uint8(124), want: true},
		{name: "float, both nil", t: enums.FloatType},
		{name: "float, first nil", t: enums.FloatType, v2: "1.23", want: true},
		{name: "float, second nil", t: enums.FloatType, v1: "0.1"},
		{name: "float, not lesser", t: enums.FloatType, v1: "122.5", v2: 122.4999},
		{name: "float, lesser", t: enums.FloatType, v1: "123", v2: float32(123.1), want: true},
		{name: "bool, both nil", t: enums.BooleanType},
		{name: "bool, first nil", t: enums.BooleanType, v2: "true", want: true},
		{name: "bool, second nil", t: enums.BooleanType, v1: "0.0"},
		{name: "bool, not lesser", t: enums.BooleanType, v1: "true"},
		{name: "bool, lesser", t: enums.BooleanType, v1: "false", v2: 1.0, want: true},
		{name: "date, both nil", t: enums.DateType},
		{name: "date, first nil", t: enums.DateType, v2: "2025-01-01", want: true},
		{name: "date, second nil", t: enums.DateType, v1: "2025-01-01"},
		{name: "date, lesser", t: enums.DateType, v1: "2025-01-01", v2: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), want: true},
		{name: "date, not lesser", t: enums.DateType, v1: "2025-01-01", v2: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{name: "any, not lesser", t: enums.AnyType, v1: "1", v2: 1},
		{name: "any, lesser", t: enums.AnyType, v1: 1.23, v2: 1.230001, want: true},
		{name: "struct, not lesser", t: enums.ObjectType, v1: s1, v2: s3},
		{name: "struct, lesser", t: enums.ObjectType, v1: s2, v2: s1, want: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Lesser(tc.t, tc.insensitive, tc.v1, tc.v2)
			assert.Equal(t, tc.want, got)
		})
	}
}
