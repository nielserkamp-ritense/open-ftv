package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyToUUID2(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want string
	}{
		{
			name: "nil",
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "string",
			in:   "abc",
			want: "abc",
		},
		{
			name: "empty byte array",
			in:   []byte{},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "filled byte array",
			in:   []byte{1, 2, 3, 4},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "wrong length",
			in:   [15]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "good length - all zero",
			in:   [16]uint8{},
			want: "00000000-0000-0000-0000-000000000000",
		},
		{
			name: "good",
			in:   [16]uint8{180, 145, 17, 36, 233, 42, 72, 47, 128, 180, 235, 2, 56, 51, 120, 174},
			want: "b4911124-e92a-482f-80b4-eb02383378ae",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AnyToUUID(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
