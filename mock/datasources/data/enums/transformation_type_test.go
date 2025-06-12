package enums

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformationTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want TransformationType
	}{
		{name: "x", want: TransformCompare},
		{name: "bad name", want: TransformCompare},
		{name: "compare", want: TransformCompare},
		{name: "convert", want: TransformConvert},
		{name: "age", want: TransformAge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := TransformationTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTransformationType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    TransformationType
		want string
	}{
		{name: "unknown", t: 199, want: "[Unknown]"},
		{name: "compare", t: TransformCompare, want: "Compare"},
		{name: "convert", t: TransformConvert, want: "Convert"},
		{name: "age", t: TransformAge, want: "Age"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTransformationType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    TransformationType
	}{
		{name: "compare", t: TransformCompare},
		{name: "convert", t: TransformConvert},
		{name: "age", t: TransformAge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 TransformationType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestTransformationType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    TransformationType
	}{
		{name: "compare", t: TransformCompare},
		{name: "convert", t: TransformConvert},
		{name: "age", t: TransformAge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 TransformationType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestTransformationType_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		f := TransformationType(199)

		got, err := f.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, `"?invalid?"`, string(got))

		got, err = f.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, "?invalid?", string(got))
	})
}
