package compare

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
)

func TestEqual(t *testing.T) {
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
		{name: "string, both nil", t: enums.StringType, want: true},
		{name: "string, first nil", t: enums.StringType, v2: "haha"},
		{name: "string, second nil", t: enums.StringType, v1: "haha"},
		{name: "string, not Equal", t: enums.StringType, v1: "hihi", v2: "Hihi"},
		{name: "string, Equal", t: enums.StringType, v1: "hoho", v2: "hoho", want: true},
		{name: "string, equalFold", t: enums.StringType, insensitive: true, v1: "hOhO", v2: "HoHo", want: true},
		{name: "integer, both nil", t: enums.IntegerType, want: true},
		{name: "integer, first nil", t: enums.IntegerType, v2: "-1"},
		{name: "integer, second nil", t: enums.IntegerType, v1: "0", want: true},
		{name: "integer, not Equal", t: enums.IntegerType, v1: "123", v2: uint8(124)},
		{name: "integer, Equal", t: enums.IntegerType, v1: "-123", v2: -123.0, want: true},
		{name: "unsigned, both nil", t: enums.UnsignedIntegerType, want: true},
		{name: "unsigned, first nil", t: enums.UnsignedIntegerType, v2: "1"},
		{name: "unsigned, second nil", t: enums.UnsignedIntegerType, v1: "0", want: true},
		{name: "unsigned, not Equal", t: enums.UnsignedIntegerType, v1: "123", v2: 122.4999},
		{name: "unsigned, Equal", t: enums.UnsignedIntegerType, v1: "123", v2: uint8(123), want: true},
		{name: "float, both nil", t: enums.FloatType, want: true},
		{name: "float, first nil", t: enums.FloatType, v2: "1.23"},
		{name: "float, second nil", t: enums.FloatType, v1: "0.0", want: true},
		{name: "float, not Equal", t: enums.FloatType, v1: "122.5", v2: 122.4999},
		{name: "float, Equal", t: enums.FloatType, v1: "123", v2: float32(123), want: true},
		{name: "bool, both nil", t: enums.BooleanType, want: true},
		{name: "bool, first nil", t: enums.BooleanType, v2: "true"},
		{name: "bool, second nil", t: enums.BooleanType, v1: "0.0", want: true},
		{name: "bool, not Equal", t: enums.BooleanType, v1: "false", v2: true},
		{name: "bool, Equal", t: enums.BooleanType, v1: "true", v2: 1.0, want: true},
		{name: "date, both nil", t: enums.DateType, want: true},
		{name: "date, first nil", t: enums.DateType, v2: "2025-01-01"},
		{name: "date, second nil", t: enums.DateType, v1: "2025-01-01"},
		{name: "date, not Equal", t: enums.DateType, v1: "2025-01-01", v2: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
		{name: "date, Equal", t: enums.DateType, v1: "2025-01-01", v2: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), want: true},
		{name: "time, both nil", t: enums.TimeType, want: true},
		{name: "time, first nil", t: enums.TimeType, v2: "13:14:15"},
		{name: "time, second nil", t: enums.TimeType, v1: "13:14:15"},
		{name: "time, not Equal", t: enums.TimeType, v1: "13:14:15", v2: time.Date(0, 1, 1, 13, 14, 15, 1, time.UTC)},
		{name: "time, Equal", t: enums.TimeType, v1: "13:14:15", v2: time.Date(0, 1, 1, 13, 14, 15, 0, time.UTC), want: true},
		{name: "date+time, both nil", t: enums.DateTimeType, want: true},
		{name: "date+time, first nil", t: enums.DateTimeType, v2: "2025-01-01 13:14:15.123"},
		{name: "date+time, second nil", t: enums.DateTimeType, v1: "2025-01-01 13:14:15.123"},
		{name: "date+time, not Equal", t: enums.DateTimeType, v1: "2025-01-01 13:14:15.123", v2: time.Date(2025, 1, 1, 13, 14, 15, 123000001, time.UTC)},
		{name: "date+time, Equal", t: enums.DateTimeType, v1: "2025-01-01 13:14:15.123", v2: time.Date(2025, 1, 1, 13, 14, 15, 123000000, time.UTC), want: true},
		{name: "any, not Equal", t: enums.AnyType, v1: "1", v2: 1},
		{name: "any, Equal", t: enums.AnyType, v1: 1.23, v2: 1.23, want: true},
		{name: "struct, not Equal", t: enums.ObjectType, v1: s1, v2: s2},
		{name: "struct, Equal", t: enums.ObjectType, v1: s1, v2: s3, want: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Equal(tc.t, tc.insensitive, tc.v1, tc.v2)
			assert.Equal(t, tc.want, got)
		})
	}
}
