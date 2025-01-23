package xsd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToDurationString(t *testing.T) {
	testCases := []struct {
		name    string
		in      any
		want    string
		wantErr bool
	}{
		{name: "nil", wantErr: true},
		{name: "bad string", in: "haha", wantErr: true},
		{name: "duration string", in: "PT4H10M", want: "PT4H10M"},
		{name: "zero", in: time.Duration(0), want: "PT0S"},
		{name: "nanoseconds", in: time.Duration(123), want: "PT0.000000123S"},
		{name: "microseconds", in: 44 * time.Microsecond, want: "PT0.000044S"},
		{name: "milliseconds", in: 12340 * time.Millisecond, want: "PT12.34S"},
		{name: "seconds", in: time.Second * 41, want: "PT41S"},
		{name: "seconds negative", in: time.Second * -41, want: "-PT41S"},
		{name: "minutes", in: time.Minute * 42, want: "PT42M"},
		{name: "minutes negative", in: time.Minute * -42, want: "-PT42M"},
		{name: "hours", in: time.Hour * 43, want: "PT43H"},
		{name: "hours negative", in: time.Hour * -43, want: "-PT43H"},
		{name: "mixed", in: time.Hour*3 + time.Minute*9 + time.Millisecond*43769, want: "PT3H9M43.769S"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toDurationString(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestToString(t *testing.T) {
	testCases := []struct {
		name    string
		in      any
		t       string
		want    string
		wantErr bool
	}{
		{name: "bad type", t: "xsd:invalid", wantErr: true},
		{name: "any", in: int16(1234), t: "xsd:anyType", want: "1234"},
		{name: "string", in: "hello world", t: "xsd:string", want: "hello world"},
		{name: "bool true", in: true, t: "xsd:boolean", want: "true"},
		{name: "bool string 1", in: "1", t: "xsd:boolean", want: "true"},
		{name: "float", in: 12.34, t: "xsd:float", want: "12.34"},
		{name: "double", in: 1234.56789, t: "xsd:double", want: "1234.56789"},
		{name: "long", in: -1234567890, t: "xsd:long", want: "-1234567890"},
		{name: "unsigned byte", in: uint8(42), t: "xsd:unsignedByte", want: "42"},
		{name: "duration", in: time.Millisecond * 1234567, t: "xsd:duration", want: "PT20M34.567S"},
		{name: "date", in: time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local), t: "xsd:date", want: "2024-12-31"},
		{name: "time", in: time.Date(0, 1, 1, 13, 14, 15, 678000000, time.Local), t: "xsd:time", want: "13:14:15.678"},
		{name: "datetime", in: time.Date(2024, 12, 31, 14, 15, 16, 780000000, time.UTC), t: "xsd:dateTime", want: "2024-12-31T14:15:16.78Z"},
		{name: "yearMonth", in: time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local), t: "xsd:gYearMonth", want: "2024-12"},
		{name: "monthDay", in: time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local), t: "xsd:gMonthDay", want: "12-31"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ToString(tc.in, tc.t)
			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
