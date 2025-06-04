package compare

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
)

func TestGreater(t *testing.T) {
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
		{name: "bad type", t: 250},
		{name: "string, both nil", t: enums.StringType},
		{name: "string, first nil", t: enums.StringType, v2: "haha"},
		{name: "string, second nil", t: enums.StringType, v1: "haha", want: true},
		{name: "string, greater", t: enums.StringType, v1: "hihi", v2: "Hihi", want: true},
		{name: "string, not greater", t: enums.StringType, v1: "hoho", v2: "hoho"},
		{name: "string, equalFold", t: enums.StringType, insensitive: true, v1: "hOhO", v2: "HoHo"},
		{name: "integer, both nil", t: enums.IntegerType},
		{name: "integer, first nil", t: enums.IntegerType, v2: "-1", want: true},
		{name: "integer, second nil", t: enums.IntegerType, v1: "0"},
		{name: "integer, not greater", t: enums.IntegerType, v1: "123", v2: uint8(124)},
		{name: "integer, greater", t: enums.IntegerType, v1: "-122", v2: -123.0, want: true},
		{name: "unsigned, both nil", t: enums.UnsignedIntegerType},
		{name: "unsigned, first nil", t: enums.UnsignedIntegerType, v2: "0"},
		{name: "unsigned, second nil", t: enums.UnsignedIntegerType, v1: "1", want: true},
		{name: "unsigned, greater", t: enums.UnsignedIntegerType, v1: "123", v2: 122.4999, want: true},
		{name: "unsigned, not greater", t: enums.UnsignedIntegerType, v1: "123", v2: uint8(124)},
		{name: "float, both nil", t: enums.FloatType},
		{name: "float, first nil", t: enums.FloatType, v2: "1.23"},
		{name: "float, second nil", t: enums.FloatType, v1: "0.1", want: true},
		{name: "float, greater", t: enums.FloatType, v1: "122.5", v2: 122.4999, want: true},
		{name: "float, not greater", t: enums.FloatType, v1: "123", v2: float32(123.1)},
		{name: "bool, both nil", t: enums.BooleanType},
		{name: "bool, first nil", t: enums.BooleanType, v2: "true"},
		{name: "bool, second nil", t: enums.BooleanType, v1: "1", want: true},
		{name: "bool, not greater", t: enums.BooleanType, v1: "false"},
		{name: "bool, greater", t: enums.BooleanType, v1: "true", v2: 0.0, want: true},
		{name: "date, both nil", t: enums.DateType},
		{name: "date, first nil", t: enums.DateType, v2: "2025-01-01"},
		{name: "date, second nil", t: enums.DateType, v1: "2025-01-01", want: true},
		{name: "date, not greater", t: enums.DateType, v1: "2025-01-01", v2: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
		{name: "date, greater", t: enums.DateType, v1: "2025-01-02", v2: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), want: true},
		{name: "any, not greater", t: enums.AnyType, v1: "1", v2: 1},
		{name: "any, greater", t: enums.AnyType, v1: 1.230001, v2: 1.23, want: true},
		{name: "struct, greater", t: enums.ObjectType, v1: s1, v2: s2, want: true},
		{name: "struct, not greater", t: enums.ObjectType, v1: s1, v2: s3},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Greater(tc.t, tc.insensitive, tc.v1, tc.v2)
			assert.Equal(t, tc.want, got)
		})
	}
}
