package convert

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToYYYYMMDD(t *testing.T) {
	var (
		tz, _ = time.LoadLocation("Europe/Amsterdam")
		d1    = time.Date(2024, 10, 1, 0, 0, 0, 0, tz)
		d2    = time.Date(2001, 1, 9, 0, 0, 0, 0, tz)
	)

	testCases := []struct {
		date *time.Time
		want string
	}{
		{},
		{date: &time.Time{}},
		{date: &d1, want: "20241001"},
		{date: &d2, want: "20010109"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.date), func(t *testing.T) {
			got := ToYYYYMMDD(tc.date)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMustYYYYMMDD(t *testing.T) {
	var d1 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		input string
		want  *time.Time
	}{
		{},
		{input: "x"},
		{input: "240101"},
		{input: "20240101", want: &d1},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := MustYYYYMMDD(tc.input)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
