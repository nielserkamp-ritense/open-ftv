package matching

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFieldMatcher(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		exp        string
		match      string
		wantAlways bool
		wantExact  []string
		wantRX     int
		want       bool
	}{
		{
			name:       "empty",
			match:      "abc",
			wantAlways: true,
			want:       true,
		},
		{
			name:       "all",
			exp:        "*",
			match:      "x",
			wantAlways: true,
			want:       true,
		},
		{
			name:      "single exact",
			exp:       "f1",
			match:     "f11",
			wantExact: []string{"f1"},
		},
		{
			name:   "single rx",
			exp:    "f*",
			match:  "f9999",
			wantRX: 1,
			want:   true,
		},
		{
			name:      "multiple exact",
			exp:       "f1,f2,f6",
			match:     "f6",
			wantExact: []string{"f1", "f2", "f6"},
			want:      true,
		},
		{
			name:   "multiple rx",
			exp:    "b*,a*,f*",
			match:  "f5",
			wantRX: 3,
			want:   true,
		},
		{
			name:      "mixed exact & rx",
			exp:       "f4,b*,f5,c??",
			match:     "cat",
			wantExact: []string{"f4", "f5"},
			wantRX:    2,
			want:      true,
		},
		{
			name:       "mixed exact, rx, empties & all",
			exp:        "f*,c??,f3,f4,,,*",
			match:      "abc",
			wantAlways: true,
			want:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m1 := NewFieldMatcher(tc.exp)
			require.NotNil(t, m1)
			assert.Equal(t, tc.wantAlways, m1.Always())

			m2, ok := m1.(*matcher)
			require.True(t, ok)
			require.NotNil(t, m2)
			assert.EqualValues(t, tc.wantExact, m2.exact)
			assert.Equal(t, tc.wantRX, len(m2.rx))

			got := m2.Match(tc.match)
			assert.Equal(t, tc.want, got)
		})
	}
}
