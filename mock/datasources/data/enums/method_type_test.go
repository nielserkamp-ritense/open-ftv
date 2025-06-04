package enums

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want MethodType
	}{
		{name: "x", want: GetMethod},
		{name: "bad name", want: GetMethod},
		{name: "get", want: GetMethod},
		{name: "post", want: PostMethod},
		{name: "put", want: PutMethod},
		{name: "delete", want: DeleteMethod},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := MethodTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMethodType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    MethodType
		want string
	}{
		{name: "unknown", t: 199, want: "[Unknown]"},
		{name: "get", t: GetMethod, want: "GET"},
		{name: "post", t: PostMethod, want: "POST"},
		{name: "put", t: PutMethod, want: "PUT"},
		{name: "delete", t: DeleteMethod, want: "DELETE"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMethodType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    MethodType
	}{
		{name: "get", t: GetMethod},
		{name: "post", t: PostMethod},
		{name: "put", t: PutMethod},
		{name: "delete", t: DeleteMethod},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 MethodType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestMethodType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    MethodType
	}{
		{name: "get", t: GetMethod},
		{name: "post", t: PostMethod},
		{name: "put", t: PutMethod},
		{name: "delete", t: DeleteMethod},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 MethodType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestMethodType_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		f := MethodType(199)

		got, err := f.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, `"?invalid?"`, string(got))

		got, err = f.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, "?invalid?", string(got))
	})
}
