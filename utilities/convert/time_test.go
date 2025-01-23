package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAnyToDateTime(t *testing.T) {
	d1 := time.Date(2024, 12, 31, 16, 17, 18, 991000000, time.UTC)

	testCases := []struct {
		name string
		in   any
		want time.Time
	}{
		{
			name: "time.Time",
			in:   d1,
			want: d1,
		},
		{
			name: "string bad",
			in:   "haha",
			want: time.Time{},
		},
		{
			name: "struct{}",
			in:   struct{}{},
			want: time.Time{},
		},
		{
			name: "double",
			in:   20241231.161718,
			want: time.Time{},
		},
		{
			name: "string date+time",
			in:   "2024-12-31T16:17:18.991Z",
			want: d1,
		},
		{
			name: "string date",
			in:   "2024-12-31",
			want: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "string time",
			in:   "16:17:18.991",
			want: time.Date(0, 1, 1, 16, 17, 18, 991000000, time.UTC),
		},
		{
			name: "string year+month",
			in:   "2024-12",
			want: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "string month+day",
			in:   "12-31",
			want: time.Date(0, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := AnyToDateTime(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
