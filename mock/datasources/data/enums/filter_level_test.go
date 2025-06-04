package enums

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterLevelFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want FilterLevel
	}{
		{name: "x", want: PrimaryLevel},
		{name: "bad name", want: PrimaryLevel},
		{name: "primary", want: PrimaryLevel},
		{name: "PrimaryLevel", want: PrimaryLevel},
		{name: "allLevel", want: AnyLevel},
		{name: "all", want: AnyLevel},
		{name: "any", want: AnyLevel},
		{name: "ANYLEVEL", want: AnyLevel},
		{name: "Join", want: JoinLevel},
		{name: "joinLevel", want: JoinLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterLevelFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFilterLevel_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FilterLevel
		want string
	}{
		{name: "unknown", t: 99, want: "[unknown]"},
		{name: "primary", t: PrimaryLevel, want: "primary"},
		{name: "any", t: AnyLevel, want: "any"},
		{name: "join", t: JoinLevel, want: "join"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFilterLevel_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FilterLevel
	}{
		{name: "primary", t: PrimaryLevel},
		{name: "any", t: AnyLevel},
		{name: "join", t: JoinLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 FilterLevel
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestFilterLevel_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    FilterLevel
	}{
		{name: "primary", t: PrimaryLevel},
		{name: "any", t: AnyLevel},
		{name: "join", t: JoinLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 FilterLevel
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}
