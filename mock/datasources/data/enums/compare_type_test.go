package enums

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want CompareType
	}{
		{name: "x"},
		{name: "bad name"},
		{name: "==", want: IsEqual},
		{name: "equal", want: IsEqual},
		{name: "not equal", want: IsNotEqual},
		{name: "!=", want: IsNotEqual},
		{name: "exists", want: Exists},
		{name: "not exists", want: NotExists},
		{name: "nil", want: NotExists},
		{name: "empty", want: NotExists},
		{name: "not empty", want: Exists},
		{name: "not nil", want: Exists},
		{name: "smaller", want: IsLesser},
		{name: "<", want: IsLesser},
		{name: "smaller or equal", want: IsLesserOrEqual},
		{name: "<=", want: IsLesserOrEqual},
		{name: "greater", want: IsGreater},
		{name: ">", want: IsGreater},
		{name: "greater or equal", want: IsGreaterOrEqual},
		{name: ">=", want: IsGreaterOrEqual},
		{name: "in list", want: InList},
		{name: "not in list", want: NotInList},
		{name: "like", want: IsLike},
		{name: "not like", want: IsNotLike},
		{name: "wc", want: IsLike},
		{name: "not wildcard", want: IsNotLike},
		{name: "match regex", want: MatchRegex},
		{name: "~", want: MatchRegex},
		{name: "!~", want: NotMatchRegex},
		{name: "not match regex", want: NotMatchRegex},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := CompareTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCompareType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    CompareType
		want string
	}{
		{name: "unknown", t: 99, want: "[unknown]"},
		{name: "equal", t: IsEqual, want: "IsEqual"},
		{name: "not equal", t: IsNotEqual, want: "IsNotEqual"},
		{name: "exists", t: Exists, want: "Exists"},
		{name: "not exists", t: NotExists, want: "NotExists"},
		{name: "smaller", t: IsLesser, want: "IsLesser"},
		{name: "smaller/equal", t: IsLesserOrEqual, want: "IsLesserOrEqual"},
		{name: "greater", t: IsGreater, want: "IsGreater"},
		{name: "greater/equal", t: IsGreaterOrEqual, want: "IsGreaterOrEqual"},
		{name: "in list", t: InList, want: "InList"},
		{name: "not in list", t: NotInList, want: "NotInList"},
		{name: "like", t: IsLike, want: "IsLike"},
		{name: "not like", t: IsNotLike, want: "IsNotLike"},
		{name: "match rx", t: MatchRegex, want: "MatchRegex"},
		{name: "not match rx", t: NotMatchRegex, want: "NotMatchRegex"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCompareType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    CompareType
	}{
		{name: "equal", t: IsEqual},
		{name: "not equal", t: IsNotEqual},
		{name: "exists", t: Exists},
		{name: "not exists", t: NotExists},
		{name: "lesser", t: IsLesser},
		{name: "lesser/equal", t: IsLesserOrEqual},
		{name: "greater", t: IsGreater},
		{name: "greater/ equal", t: IsGreaterOrEqual},
		{name: "in list", t: InList},
		{name: "not in list", t: NotInList},
		{name: "like", t: IsLike},
		{name: "not like", t: IsNotLike},
		{name: "regex", t: MatchRegex},
		{name: "not regex", t: NotMatchRegex},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 CompareType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
			assert.True(t, t2.IsValid())
		})
	}
}

func TestCompareType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    CompareType
	}{
		{name: "equal", t: IsEqual},
		{name: "not equal", t: IsNotEqual},
		{name: "exists", t: Exists},
		{name: "not exists", t: NotExists},
		{name: "lesser", t: IsLesser},
		{name: "lesser/equal", t: IsLesserOrEqual},
		{name: "greater", t: IsGreater},
		{name: "greater/ equal", t: IsGreaterOrEqual},
		{name: "in list", t: InList},
		{name: "not in list", t: NotInList},
		{name: "like", t: IsLike},
		{name: "not like", t: IsNotLike},
		{name: "regex", t: MatchRegex},
		{name: "not regex", t: NotMatchRegex},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 CompareType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
			assert.True(t, t2.IsValid())
		})
	}
}
