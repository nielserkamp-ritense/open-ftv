package types

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinTypeFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want JoinType
	}{
		{name: "x", want: OptionalParentChild},
		{name: "bad name", want: OptionalParentChild},
		{name: "parent", want: OptionalParentChild},
		{name: "child", want: OptionalParentChild},
		{name: "parentchild", want: OptionalParentChild},
		{name: "sibling", want: OptionalSibling},
		{name: "optionalparentchild", want: OptionalParentChild},
		{name: "optionalsibling", want: OptionalSibling},
		{name: "forced", want: ForcedParentChild},
		{name: "forcedparent", want: ForcedParentChild},
		{name: "forcedchild", want: ForcedParentChild},
		{name: "forcedparentchild", want: ForcedParentChild},
		{name: "forcedsibling", want: ForcedSibling},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := JoinTypeFromString(tc.name)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestJoinType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    JoinType
		want string
	}{
		{name: "unknown", t: 199, want: "[unknown]"},
		{name: "optional parent/child", t: OptionalParentChild, want: "OptionalParentChild"},
		{name: "optional sibling", t: OptionalSibling, want: "OptionalSibling"},
		{name: "forced parent/child", t: ForcedParentChild, want: "ForcedParentChild"},
		{name: "forced sibling", t: ForcedSibling, want: "ForcedSibling"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestJoinType_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    JoinType
	}{
		{name: "optional parent/child", t: OptionalParentChild},
		{name: "optional sibling", t: OptionalSibling},
		{name: "forced parent/child", t: ForcedParentChild},
		{name: "forced sibling", t: ForcedSibling},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 JoinType
			err = json.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestJoinType_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		t    JoinType
	}{
		{name: "optional parent/child", t: OptionalParentChild},
		{name: "optional sibling", t: OptionalSibling},
		{name: "forced parent/child", t: ForcedParentChild},
		{name: "forced sibling", t: ForcedSibling},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.t)
			require.NoError(t, err)
			require.NotNil(t, got)

			var t2 JoinType
			err = yaml.Unmarshal(got, &t2)
			require.NoError(t, err)
			assert.Equal(t, tc.t, t2)
		})
	}
}

func TestJoinType_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("unknown type", func(t *testing.T) {
		t.Parallel()

		f := JoinType(199)

		got, err := f.MarshalJSON()
		require.NoError(t, err)
		assert.Equal(t, `"[unknown]"`, string(got))

		got, err = f.MarshalYAML()
		require.NoError(t, err)
		assert.Equal(t, "[unknown]", string(got))
	})
}
