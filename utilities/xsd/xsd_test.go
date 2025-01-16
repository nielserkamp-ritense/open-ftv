package xsd

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertXSD(t *testing.T) {
	u1, _ := url.Parse("https://ftv.nl/helloWorld#second?format=text")

	testCases := []struct {
		name    string
		data    string
		t       string
		want    any
		wantErr bool
	}{
		{name: "no mime-type", data: "abc", wantErr: true},
		{name: "invalid mime-type", data: "abc", t: "abc", wantErr: true},
		{name: "string", data: "abc", t: "http://www.w3.org/2001/XMLSchema#string", want: "abc"},
		{name: "any", data: "abc", t: "http://www.w3.org/2001/XMLSchema#anyType", want: "abc"},
		{name: "simple", data: "abc", t: "http://www.w3.org/2001/XMLSchema#simpleType", want: "abc"},
		{name: "normalized string", data: "abc", t: "http://www.w3.org/2001/XMLSchema#normalizedString", want: "abc"},
		{name: "XSD token", data: "abc", t: "http://www.w3.org/2001/XMLSchema#token", want: "abc"},
		{name: "XSD language", data: "abc", t: "http://www.w3.org/2001/XMLSchema#language", want: "abc"},
		{name: "boolean true", data: "true", t: "http://www.w3.org/2001/XMLSchema#boolean", want: true},
		{name: "boolean false", data: "false", t: "http://www.w3.org/2001/XMLSchema#boolean", want: false},
		{name: "boolean 1", data: "1", t: "http://www.w3.org/2001/XMLSchema#boolean", want: true},
		{name: "boolean 0", data: "0", t: "http://www.w3.org/2001/XMLSchema#boolean", want: false},
		{name: "float", data: "1.25", t: "http://www.w3.org/2001/XMLSchema#float", want: 1.25},
		{name: "double", data: "1.23456", t: "http://www.w3.org/2001/XMLSchema#double", want: 1.23456},
		{name: "decimal with fraction", data: "1.23", t: "http://www.w3.org/2001/XMLSchema#decimal", want: 1.23},
		{name: "decimal without fraction", data: "123", t: "http://www.w3.org/2001/XMLSchema#decimal", want: int64(123)},
		{name: "decimal error", data: "xyz", t: "http://www.w3.org/2001/XMLSchema#decimal", wantErr: true},
		{name: "integer good", data: "98765", t: "http://www.w3.org/2001/XMLSchema#integer", want: int64(98765)},
		{name: "integer error", data: "xyz", t: "http://www.w3.org/2001/XMLSchema#integer", wantErr: true},
		{name: "long good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#long", want: int64(987654321)},
		{name: "long error", data: "00xyz", t: "http://www.w3.org/2001/XMLSchema#long", wantErr: true},
		{name: "int good", data: "9876543", t: "http://www.w3.org/2001/XMLSchema#int", want: int64(9876543)},
		{name: "int error", data: "15x99", t: "http://www.w3.org/2001/XMLSchema#int", wantErr: true},
		{name: "short good", data: "8765", t: "http://www.w3.org/2001/XMLSchema#short", want: int64(8765)},
		{name: "short error", data: "32769", t: "http://www.w3.org/2001/XMLSchema#short", wantErr: true},
		{name: "byte good", data: "65", t: "http://www.w3.org/2001/XMLSchema#byte", want: int64(65)},
		{name: "byte error", data: "129", t: "http://www.w3.org/2001/XMLSchema#byte", wantErr: true},
		{name: "ulong good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#unsignedLong", want: uint64(987654321)},
		{name: "ulong error", data: "-1", t: "http://www.w3.org/2001/XMLSchema#unsignedLong", wantErr: true},
		{name: "uint good", data: "9876543", t: "http://www.w3.org/2001/XMLSchema#unsignedInt", want: uint64(9876543)},
		{name: "uint error", data: "-2", t: "http://www.w3.org/2001/XMLSchema#unsignedInt", wantErr: true},
		{name: "ushort good", data: "8765", t: "http://www.w3.org/2001/XMLSchema#unsignedShort", want: uint64(8765)},
		{name: "ushort error", data: "65536", t: "http://www.w3.org/2001/XMLSchema#unsignedShort", wantErr: true},
		{name: "ubyte good", data: "65", t: "http://www.w3.org/2001/XMLSchema#unsignedByte", want: uint64(65)},
		{name: "ubyte error", data: "256", t: "http://www.w3.org/2001/XMLSchema#unsignedByte", wantErr: true},
		{name: "nonPositive good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#nonPositiveInteger", want: int64(987654321)},
		{name: "nonPositive error", data: "+p", t: "http://www.w3.org/2001/XMLSchema#nonPositiveInteger", wantErr: true},
		{name: "negative good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#negativeInteger", want: int64(987654321)},
		{name: "negative error", data: "q", t: "http://www.w3.org/2001/XMLSchema#negativeInteger", wantErr: true},
		{name: "nonNegative good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#nonNegativeInteger", want: uint64(987654321)},
		{name: "nonNegative error", data: "-4", t: "http://www.w3.org/2001/XMLSchema#nonNegativeInteger", wantErr: true},
		{name: "positive good", data: "987654321", t: "http://www.w3.org/2001/XMLSchema#positiveInteger", want: uint64(987654321)},
		{name: "positive error", data: "-1", t: "http://www.w3.org/2001/XMLSchema#positiveInteger", wantErr: true},
		{name: "year good", data: "9876543", t: "http://www.w3.org/2001/XMLSchema#gYear", want: uint64(9876543)},
		{name: "year error", data: "-y", t: "http://www.w3.org/2001/XMLSchema#gYear", wantErr: true},
		{name: "month good", data: "99", t: "http://www.w3.org/2001/XMLSchema#gMonth", want: uint64(99)},
		{name: "month error", data: "256", t: "http://www.w3.org/2001/XMLSchema#gMonth", wantErr: true},
		{name: "day good", data: "99", t: "http://www.w3.org/2001/XMLSchema#gDay", want: uint64(99)},
		{name: "day error", data: "256", t: "http://www.w3.org/2001/XMLSchema#gDay", wantErr: true},
		{name: "uri good", data: "https://ftv.nl/helloWorld#second?format=text", t: "http://www.w3.org/2001/XMLSchema#anyURI", want: u1},
		{name: "uri error", data: "https://ftv.nl/hell\000World#second?format=text", t: "http://www.w3.org/2001/XMLSchema#anyURI", wantErr: true},
		{name: "duration good", data: "PT25S", t: "http://www.w3.org/2001/XMLSchema#duration", want: 25 * time.Second},
		{name: "duration error (1)", data: "abc", t: "http://www.w3.org/2001/XMLSchema#duration", wantErr: true},
		{name: "duration error (2)", data: "PT1.2.3S", t: "http://www.w3.org/2001/XMLSchema#duration", wantErr: true},
		{name: "date+time", data: "2024-12-28T13:29:15Z", t: "http://www.w3.org/2001/XMLSchema#dateTime", want: time.Date(2024, 12, 28, 13, 29, 15, 0, time.UTC)},
		{name: "date+time error", data: "abc", t: "http://www.w3.org/2001/XMLSchema#dateTime", wantErr: true},
		{name: "time good", data: "13:59:59.887432581", t: "http://www.w3.org/2001/XMLSchema#time", want: time.Date(0, 1, 1, 13, 59, 59, 887432581, time.Local)},
		{name: "time bad", data: "13:61:59.887432581", t: "http://www.w3.org/2001/XMLSchema#time", wantErr: true},
		{name: "date good", data: "2024-12-29", t: "http://www.w3.org/2001/XMLSchema#date", want: time.Date(2024, 12, 29, 0, 0, 0, 0, time.Local)},
		{name: "time bad", data: "2024-02-30", t: "http://www.w3.org/2001/XMLSchema#date", wantErr: true},
		{name: "year+month good", data: "2024-12", t: "http://www.w3.org/2001/XMLSchema#gYearMonth", want: time.Date(2024, 12, 1, 0, 0, 0, 0, time.Local)},
		{name: "year+month bad", data: "2024-13", t: "http://www.w3.org/2001/XMLSchema#gYearMonth", wantErr: true},
		{name: "month+day good", data: "12-31", t: "http://www.w3.org/2001/XMLSchema#gMonthDay", want: time.Date(0, 12, 31, 0, 0, 0, 0, time.Local)},
		{name: "month+day bad", data: "02-30", t: "http://www.w3.org/2001/XMLSchema#gMonthDay", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err2 := Convert(tc.data, tc.t)
			if tc.wantErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestConvertDecimal(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		want    any
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "invalid", data: "abc", wantErr: true},
		{name: "zero", data: "0", want: int64(0)},
		{name: "zero with decimals", data: "0.0", want: 0.0},
		{name: "integer", data: "123456789", want: int64(123456789)},
		{name: "float", data: "12345.6789", want: float64(12345.6789)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertDecimal(tc.data)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestConvertDuration(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		want    any
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "invalid", data: "abc", wantErr: true},
		{name: "zero", data: "PT0S", want: time.Duration(0)},
		{name: "seconds", data: "PT12S", want: 12 * time.Second},
		{name: "minutes", data: "PT23M", want: 23 * time.Minute},
		{name: "hours", data: "PT7H", want: 7 * time.Hour},
		{name: "time", data: "PT7H23M12S", want: 7*time.Hour + 23*time.Minute + 12*time.Second},
		{name: "days", data: "P4D", want: "P4D"},
		{name: "months", data: "P2M", want: "P2M"},
		{name: "years", data: "P1Y", want: "P1Y"},
		{name: "date", data: "P1Y2M4D", want: "P1Y2M4D"},
		{name: "date+time", data: "P1Y2M4DT7H23M12S", want: "P1Y2M4DT7H23M12S"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertDuration(tc.data)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestConvertURI(t *testing.T) {
	s2 := "haha\000"
	s3 := "abc"
	s4 := "https://"
	s5 := "ftv.nl"
	s6 := "https://ftv.nl/rdf/dummy.ttl"

	u1, _ := url.Parse("")
	u3, _ := url.Parse(s3)
	u4, _ := url.Parse(s4)
	u5, _ := url.Parse(s5)
	u6, _ := url.Parse(s6)

	testCases := []struct {
		name    string
		data    string
		want    any
		wantErr bool
	}{
		{name: "empty", want: u1},
		{name: "invalid", data: s2, wantErr: true},
		{name: "scheme", data: s3, want: u3},
		{name: "host", data: s4, want: u4},
		{name: "path", data: s5, want: u5},
		{name: "full URI", data: s6, want: u6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertURI(tc.data)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.want, got)
			}
		})
	}
}
