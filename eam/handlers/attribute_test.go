package handlers

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

func TestConvertFromOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *attributes.Attribute
		want *models.Attribute
	}{
		{name: "no type - string", in: &attributes.Attribute{Key: "key", Value: "value"}, want: models.NewAttribute("key", "value")},
		{name: "no type - integer", in: &attributes.Attribute{Key: "key", Value: 123}, want: models.NewAttribute("key", 123)},
		{name: "no type - bool", in: &attributes.Attribute{Key: "key", Value: true}, want: models.NewAttribute("key", true)},
		{name: "no type - double", in: &attributes.Attribute{Key: "key", Value: 22.33}, want: models.NewAttribute("key", 22.33)},
		{name: "no type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("999")}, want: models.NewAttribute("key", json.Number("999"))},
		{name: "string type - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: "string"}, want: models.NewAttributeWithType("key", "value", "string")},
		{name: "string type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "string"}, want: models.NewAttributeWithType("key", "987", "string")},
		{name: "string type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "string"}, want: models.NewAttributeWithType("key", "987", "string")},
		{name: "string type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "string"}, want: models.NewAttributeWithType("key", "777", "string")},
		{name: "string type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "string"}, want: models.NewAttributeWithType("key", "987.432", "string")},
		{name: "string type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "string"}, want: models.NewAttributeWithType("key", "false", "string")},
		{name: "string type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "string"}, want: models.NewAttributeWithType("key", "[1 2 3]", "string")},
		{name: "integer type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "integer"}, want: models.NewAttributeWithType("key", int64(0), "integer")},
		{name: "integer type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "integer"}, want: models.NewAttributeWithType("key", int64(123), "integer")},
		{name: "integer type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "integer"}, want: models.NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "integer"}, want: models.NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "integer"}, want: models.NewAttributeWithType("key", int64(777), "integer")},
		{name: "integer type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "integer"}, want: models.NewAttributeWithType("key", int64(987), "integer")},
		{name: "integer type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "integer"}, want: models.NewAttributeWithType("key", int64(1), "integer")},
		{name: "integer type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "integer"}, want: models.NewAttributeWithType("key", int64(0), "integer")},
		{name: "int type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "int"}, want: models.NewAttributeWithType("key", int64(0), "int")},
		{name: "int type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "int"}, want: models.NewAttributeWithType("key", int64(123), "int")},
		{name: "int type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "int"}, want: models.NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "int"}, want: models.NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "int"}, want: models.NewAttributeWithType("key", int64(777), "int")},
		{name: "int type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "int"}, want: models.NewAttributeWithType("key", int64(987), "int")},
		{name: "int type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "int"}, want: models.NewAttributeWithType("key", int64(1), "int")},
		{name: "int type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "int"}, want: models.NewAttributeWithType("key", int64(0), "int")},
		{name: "short type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "short"}, want: models.NewAttributeWithType("key", int64(0), "short")},
		{name: "short type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "short"}, want: models.NewAttributeWithType("key", int64(123), "short")},
		{name: "long type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "long"}, want: models.NewAttributeWithType("key", int64(987), "long")},
		{name: "long type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "long"}, want: models.NewAttributeWithType("key", int64(987), "long")},
		{name: "long type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "long"}, want: models.NewAttributeWithType("key", int64(777), "long")},
		{name: "byte type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "byte"}, want: models.NewAttributeWithType("key", int64(987), "byte")},
		{name: "byte type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "byte"}, want: models.NewAttributeWithType("key", int64(1), "byte")},
		{name: "byte type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "byte"}, want: models.NewAttributeWithType("key", int64(0), "byte")},
		{name: "boolean type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "boolean"}, want: models.NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "boolean"}, want: models.NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "boolean"}, want: models.NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - int64", in: &attributes.Attribute{Key: "key", Value: int64(1), Type: "boolean"}, want: models.NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "boolean"}, want: models.NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "boolean"}, want: models.NewAttributeWithType("key", true, "boolean")},
		{name: "boolean type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "boolean"}, want: models.NewAttributeWithType("key", false, "boolean")},
		{name: "boolean type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "boolean"}, want: models.NewAttributeWithType("key", false, "boolean")},
		{name: "bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "bool"}, want: models.NewAttributeWithType("key", false, "bool")},
		{name: "bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "bool"}, want: models.NewAttributeWithType("key", true, "bool")},
		{name: "bool type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "bool"}, want: models.NewAttributeWithType("key", true, "bool")},
		{name: "bool type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "bool"}, want: models.NewAttributeWithType("key", false, "bool")},
		{name: "bool type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "bool"}, want: models.NewAttributeWithType("key", true, "bool")},
		{name: "bool type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "bool"}, want: models.NewAttributeWithType("key", false, "bool")},
		{name: "bool type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "bool"}, want: models.NewAttributeWithType("key", false, "bool")},
		{name: "double type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "double"}, want: models.NewAttributeWithType("key", float64(0), "double")},
		{name: "double type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "double"}, want: models.NewAttributeWithType("key", 123.45678, "double")},
		{name: "double type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "double"}, want: models.NewAttributeWithType("key", float64(987), "double")},
		{name: "double type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "double"}, want: models.NewAttributeWithType("key", float64(987), "double")},
		{name: "double type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "double"}, want: models.NewAttributeWithType("key", float64(777), "double")},
		{name: "double type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "double"}, want: models.NewAttributeWithType("key", 987.432, "double")},
		{name: "double type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "double"}, want: models.NewAttributeWithType("key", float64(1), "double")},
		{name: "double type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "double"}, want: models.NewAttributeWithType("key", float64(0), "double")},
		{name: "float type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "float"}, want: models.NewAttributeWithType("key", float64(0), "float")},
		{name: "float type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "float"}, want: models.NewAttributeWithType("key", 123.45678, "float")},
		{name: "float type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "float"}, want: models.NewAttributeWithType("key", float64(987), "float")},
		{name: "float type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "float"}, want: models.NewAttributeWithType("key", float64(777), "float")},
		{name: "float type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "float"}, want: models.NewAttributeWithType("key", 987.432, "float")},
		{name: "float type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "float"}, want: models.NewAttributeWithType("key", float64(1), "float")},
		{name: "float type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "float"}, want: models.NewAttributeWithType("key", float64(0), "float")},
		{name: "date type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21", Type: "date"}, want: models.NewAttributeWithType("key", time.Date(2024, 07, 21, 0, 0, 0, 0, time.UTC), "date")},
		{name: "date type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "date type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "date"}, want: models.NewAttributeWithType("key", time.Time{}, "date")},
		{name: "time type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - string good", in: &attributes.Attribute{Key: "key", Value: "17:31:58", Type: "time"}, want: models.NewAttributeWithType("key", time.Date(0, 1, 1, 17, 31, 58, 0, time.UTC), "time")},
		{name: "time type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "time type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "time"}, want: models.NewAttributeWithType("key", time.Time{}, "time")},
		{name: "dateTime type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "datetime"}, want: models.NewAttributeWithType("key", time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC), "datetime")},
		{name: "dateTime type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "dateTime type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "datetime"}, want: models.NewAttributeWithType("key", time.Time{}, "datetime")},
		{name: "timestamp type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC), "timestamp")},
		{name: "timestamp type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "timestamp type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "timestamp"}, want: models.NewAttributeWithType("key", time.Time{}, "timestamp")},
		{name: "xsd:any - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixAny}, want: models.NewAttributeWithType("key", "value", "xsd:anyType")},
		{name: "xsd:any - int", in: &attributes.Attribute{Key: "key", Value: 654321, Type: xsd.PrefixAny}, want: models.NewAttributeWithType("key", 654321, "xsd:anyType")},
		{name: "xsd:any - double", in: &attributes.Attribute{Key: "key", Value: 555.888, Type: xsd.PrefixAny}, want: models.NewAttributeWithType("key", 555.888, "xsd:anyType")},
		{name: "xsd:any - json.number", in: &attributes.Attribute{Key: "key", Value: json.Number("765"), Type: xsd.PrefixAny}, want: models.NewAttributeWithType("key", json.Number("765"), "xsd:anyType")},
		{name: "xsd:anyURI - string", in: &attributes.Attribute{Key: "key", Value: "https://localhost:443", Type: xsd.PrefixAnyURI}, want: models.NewAttributeWithType("key", "https://localhost:443", "xsd:anyURI")},
		{name: "xsd:bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: xsd.PrefixBoolean}, want: models.NewAttributeWithType("key", false, "xsd:boolean")},
		{name: "xsd:bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: xsd.PrefixBoolean}, want: models.NewAttributeWithType("key", true, "xsd:boolean")},
		{name: "xsd:string - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixString}, want: models.NewAttributeWithType("key", "value", "xsd:string")},
		{name: "xsd:duration type - string bad", in: &attributes.Attribute{Key: "key", Value: "xyz", Type: xsd.PrefixDuration}, want: models.NewAttributeWithType("key", time.Duration(0), "xsd:duration")},
		{name: "xsd:duration type - string good (1)", in: &attributes.Attribute{Key: "key", Value: "260s", Type: xsd.PrefixDuration}, want: models.NewAttributeWithType("key", 260*time.Second, "xsd:duration")},
		{name: "xsd:duration type - string good (2)", in: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: xsd.PrefixDuration}, want: models.NewAttributeWithType("key", 260*time.Second, "xsd:duration")},
		{name: "xsd:duration type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: xsd.PrefixDuration}, want: models.NewAttributeWithType("key", time.Duration(0), "xsd:duration")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AttributeFromOAS(tc.in)
			require.NotNil(t, got)
			assert.Equal(t, tc.want.Key(), got.Key())
			assert.Equal(t, tc.want.Value(), got.Value())
			assert.Equal(t, tc.in.Value, got.Original())
			assert.Equal(t, tc.want.Type(), got.Type())
		})
	}
}

func TestConvertToOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *models.Attribute
		want *attributes.Attribute
	}{
		{name: "string", in: models.NewAttribute("key", "value"), want: &attributes.Attribute{Key: "key", Value: "value", Type: "string"}},
		{name: "int64", in: models.NewAttribute("key", int64(9876543)), want: &attributes.Attribute{Key: "key", Value: int64(9876543), Type: "long"}},
		{name: "double", in: models.NewAttribute("key", 123.5678), want: &attributes.Attribute{Key: "key", Value: 123.5678, Type: "double"}},
		{name: "bool", in: models.NewAttribute("key", true), want: &attributes.Attribute{Key: "key", Value: true, Type: "bool"}},
		{name: "slice", in: models.NewAttribute("key", []any{5, 6, 7, 8}), want: &attributes.Attribute{Key: "key", Value: []any{5, 6, 7, 8}}},
		{name: "map", in: models.NewAttribute("key", map[string]any{"hello": "world", "int": 123}), want: &attributes.Attribute{Key: "key", Value: map[string]any{"hello": "world", "int": 123}}},
		{name: "time.Time", in: models.NewAttribute("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC)), want: &attributes.Attribute{Key: "key", Value: "2024-05-29T12:00:00Z", Type: "datetime"}},
		{name: "time.Duration", in: models.NewAttribute("key", time.Duration(123456780000)), want: &attributes.Attribute{Key: "key", Value: "2m3.45678s", Type: "duration"}},
		{name: "unsupported", in: models.NewAttribute("key", []float32{1, 2, 3}), want: &attributes.Attribute{Key: "key", Value: "[1 2 3]", Type: "string"}},
		{name: "with string type", in: models.NewAttributeWithType("key", "true", "string"), want: &attributes.Attribute{Key: "key", Value: "true", Type: "string"}},
		{name: "with xsd:boolean type", in: models.NewAttributeWithType("key", true, "xsd:boolean"), want: &attributes.Attribute{Key: "key", Value: true, Type: "xsd:boolean"}},
		{name: "with xsd:gYear type", in: models.NewAttributeWithType("key", 2025, "xsd:gYear"), want: &attributes.Attribute{Key: "key", Value: 2025, Type: "xsd:gYear"}},
		{name: "with date type", in: models.NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "date"), want: &attributes.Attribute{Key: "key", Value: "2024-05-29", Type: "date"}},
		{name: "with xsd:time type", in: models.NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "xsd:time"), want: &attributes.Attribute{Key: "key", Value: "12:00:00", Type: "xsd:time"}},
		{name: "with xsd:dateTime type", in: models.NewAttributeWithType("key", time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC), "xsd:dateTime"), want: &attributes.Attribute{Key: "key", Value: "2024-05-29T12:00:00Z", Type: "xsd:dateTime"}},
		{name: "with xsd:duration type (1)", in: models.NewAttributeWithType("key", 260*time.Second, "xsd:duration"), want: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: "xsd:duration"}},
		{name: "with xsd:duration type (2)", in: models.NewAttributeWithType("key", 300*time.Millisecond, "xsd:duration"), want: &attributes.Attribute{Key: "key", Value: "PT0.3S", Type: "xsd:duration"}},
		{name: "with xsd:duration type (3)", in: models.NewAttributeWithType("key", time.Duration(1200), "xsd:duration"), want: &attributes.Attribute{Key: "key", Value: "PT0.0000012S", Type: "xsd:duration"}},
		{name: "with duration type", in: models.NewAttributeWithType("key", time.Duration(123456780000), "duration"), want: &attributes.Attribute{Key: "key", Value: "2m3.45678s", Type: "duration"}},
		{name: "with original xsd:boolean", in: models.NewOriginalAttribute("key", true, "1", "xsd:boolean"), want: &attributes.Attribute{Key: "key", Value: "1", Type: "xsd:boolean"}},
		{name: "with original xsd:duration", in: models.NewOriginalAttribute("key", 260*time.Second, "PT4M20S", "xsd:duration"), want: &attributes.Attribute{Key: "key", Value: "PT4M20S", Type: "xsd:duration"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AttributeToOAS(tc.in)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
