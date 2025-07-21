package cedar_embedded

import (
	"fmt"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueToAny(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Millisecond)

	f1, _ := cedar.NewDecimalFromFloat(456.789)
	f2, _ := cedar.NewDecimalFromFloat(123.456)

	testCases := []struct {
		name    string
		in      cedar.Value
		want    any
		wantErr bool
	}{
		{
			name:    "invalid",
			in:      cedar.EntityUID{Type: "x", ID: "y"},
			want:    `x::"y"`,
			wantErr: true,
		},
		{
			name: "nil",
		},
		{
			name: "empty string",
			in:   cedar.String(""),
			want: "",
		},
		{
			name: "string",
			in:   cedar.String("world magic"),
			want: "world magic",
		},
		{
			name: "true",
			in:   cedar.Boolean(true),
			want: true,
		},
		{
			name: "false",
			in:   cedar.Boolean(false),
			want: false,
		},
		{
			name: "long",
			in:   cedar.Long(123456),
			want: int64(123456),
		},
		{
			name: "decimal",
			in:   f1,
			want: 456.789,
		},
		{
			name: "time",
			in:   cedar.NewDatetime(now),
			want: now,
		},
		{
			name: "duration",
			in:   cedar.NewDuration(15 * time.Millisecond),
			want: 15 * time.Millisecond,
		},
		{
			name: "empty set",
			in:   cedar.NewSet(),
			want: []any{},
		},
		{
			name: "set",
			in:   cedar.NewSet(cedar.String("yo"), cedar.Boolean(true)),
			want: []any{"yo", true},
		},
		{
			name: "empty map",
			in:   cedar.NewRecord(cedar.RecordMap{}),
			want: map[string]any{},
		},
		{
			name: "map",
			in:   cedar.NewRecord(cedar.RecordMap{"do": cedar.String("it"), "float": f2}),
			want: map[string]any{"do": "it", "float": 123.456},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := valueToAny(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.EqualValues(t, tc.want, got)
			} else {
				require.NoError(t, err)

				s1, ok1 := tc.want.([]any)
				s2, ok2 := got.([]any)
				if ok1 && ok2 {
					for i := range s1 {
						var found bool
						for j := range s2 {
							if s1[i] == s2[j] {
								found = true
								break
							}
						}
						assert.Truef(t, found, fmt.Sprintf("slices mismatched; %v != %v", s1, s2))
					}
				} else {
					assert.EqualValues(t, tc.want, got)
				}
			}
		})
	}
}
