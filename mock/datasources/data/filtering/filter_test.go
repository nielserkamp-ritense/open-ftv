package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		data      string
		wantErr   bool
		wantField bool
		wantAll   bool
		wantAny   bool
	}{
		{
			name:    "bad data",
			data:    "\001\002",
			wantErr: true,
		},
		{
			name:      "field value",
			data:      `{"level":"primary"}`,
			wantField: true,
		},
		{
			name:    "all of",
			data:    `{"allOf":[{"level":"primary"},{"level":"primary"}]}`,
			wantAll: true,
		},
		{
			name:    "any of",
			data:    `{"anyOf":[{"level":"primary"},{"level":"primary"}]}`,
			wantAny: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := new(Filter)
			err := f.UnmarshalJSON([]byte(tc.data))
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, f.AllOf != nil && f.AnyOf != nil)
				assert.Equal(t, tc.wantField, f.AllOf == nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAll, f.AllOf != nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAny, f.AllOf == nil && f.AnyOf != nil)
			}
		})
	}
}

func TestFilter_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		data      string
		wantErr   bool
		wantField bool
		wantAll   bool
		wantAny   bool
	}{
		{
			name:    "bad data",
			data:    "\001\002",
			wantErr: true,
		},
		{
			name:      "field value",
			data:      `level: primary`,
			wantField: true,
		},
		{
			name: "all of",
			data: `allOf:
  - level: primary
  - level: primary
`,
			wantAll: true,
		},
		{
			name: "any of",
			data: `anyOf:
  - level: primary
  - level: primary
`,
			wantAny: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := new(Filter)
			err := f.UnmarshalYAML([]byte(tc.data))
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, f.AllOf != nil && f.AnyOf != nil)
				assert.Equal(t, tc.wantField, f.AllOf == nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAll, f.AllOf != nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAny, f.AllOf == nil && f.AnyOf != nil)
			}
		})
	}
}
