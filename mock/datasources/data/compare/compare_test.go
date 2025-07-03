package compare

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestFieldValueFilter_Compare(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		params Params
		want   bool
	}{
		{
			name:   "string - invalid compare",
			params: Params{Compare: 250, Input: "x", Value: "x"},
		},
		{
			name:   "string - not exists - nil",
			params: Params{Compare: enums.NotExists},
			want:   true,
		},
		{
			name:   "string - not exists - not nil",
			params: Params{Compare: enums.Exists},
		},
		{
			name:   "string - not exists - equal",
			params: Params{Compare: enums.IsEqual, Value: "x"},
		},
		{
			name:   "string - exists - nil",
			params: Params{Compare: enums.NotExists, Input: ""},
		},
		{
			name:   "string - exists - not nil",
			params: Params{Compare: enums.Exists, Input: ""},
			want:   true,
		},
		{
			name:   "string - exists - equal",
			params: Params{Compare: enums.IsEqual, Input: "1", Value: int8(1)},
			want:   true,
		},
		{
			name:   "string - exists - not equal",
			params: Params{Compare: enums.IsNotEqual, Input: "11", Value: int8(1)},
			want:   true,
		},
		{
			name:   "string - exists - lesser",
			params: Params{Compare: enums.IsLesser, Input: "1", Value: int8(2)},
			want:   true,
		},
		{
			name:   "string - exists - lesser or equal",
			params: Params{Compare: enums.IsLesserOrEqual, Input: "11", Value: int8(11)},
			want:   true,
		},
		{
			name:   "string - exists - greater",
			params: Params{Compare: enums.IsGreater, Input: "2", Value: int8(1)},
			want:   true,
		},
		{
			name:   "string - greater or equal",
			params: Params{Compare: enums.IsGreaterOrEqual, Input: "11", Value: int8(1)},
			want:   true,
		},
		{
			name:   "integer - exists - in list",
			params: Params{Compare: enums.InList, Input: "1", Values: []any{0, 5, 112, 1, 4}},
			want:   true,
		},
		{
			name:   "integer - exists - not in list",
			params: Params{Compare: enums.NotInList, Input: "11", Values: []any{5, 6, 7, 8, 9, 10}},
			want:   true,
		},
		{
			name:   "string - exists - match regex",
			params: Params{Compare: enums.MatchRegex, Input: "hello world", RX: regexp.MustCompile(`^.*lo.?wo.*$`)},
			want:   true,
		},
		{
			name:   "string - exists - not match regex",
			params: Params{Compare: enums.NotMatchRegex, Input: "hello saturn", RX: regexp.MustCompile(`^.*lo.?ju.*$`)},
			want:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Compare(tc.params)
			assert.Equal(t, tc.want, got)
		})
	}
}
