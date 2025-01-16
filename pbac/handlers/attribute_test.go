package handlers

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

func TestConvertFromOAS(t *testing.T) {
	testCases := []struct {
		name string
		in   *attributes.Attribute
		want *models.Attribute
	}{
		{name: "no type - string", in: &attributes.Attribute{Key: "key", Value: "value"}, want: &models.Attribute{Key: "key", Value: "value"}},
		{name: "no type - integer", in: &attributes.Attribute{Key: "key", Value: 123}, want: &models.Attribute{Key: "key", Value: 123}},
		{name: "no type - bool", in: &attributes.Attribute{Key: "key", Value: true}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "no type - double", in: &attributes.Attribute{Key: "key", Value: 22.33}, want: &models.Attribute{Key: "key", Value: 22.33}},
		{name: "no type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("999")}, want: &models.Attribute{Key: "key", Value: json.Number("999")}},
		{name: "string type - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: "string"}, want: &models.Attribute{Key: "key", Value: "value"}},
		{name: "string type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "string"}, want: &models.Attribute{Key: "key", Value: "987"}},
		{name: "string type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "string"}, want: &models.Attribute{Key: "key", Value: "987"}},
		{name: "string type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "string"}, want: &models.Attribute{Key: "key", Value: "777"}},
		{name: "string type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "string"}, want: &models.Attribute{Key: "key", Value: "987.432"}},
		{name: "string type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "string"}, want: &models.Attribute{Key: "key", Value: "false"}},
		{name: "string type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "string"}, want: &models.Attribute{Key: "key", Value: "[1 2 3]"}},
		{name: "integer type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "integer type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(123)}},
		{name: "integer type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "integer type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "integer type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(777)}},
		{name: "integer type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "integer type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(1)}},
		{name: "integer type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "integer"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "int type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "int type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(123)}},
		{name: "int type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "int type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "int type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(777)}},
		{name: "int type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "int type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(1)}},
		{name: "int type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "int"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "short type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "short"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "short type - string good", in: &attributes.Attribute{Key: "key", Value: "123", Type: "short"}, want: &models.Attribute{Key: "key", Value: int64(123)}},
		{name: "long type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "long"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "long type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "long"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "long type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "long"}, want: &models.Attribute{Key: "key", Value: int64(777)}},
		{name: "byte type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "byte"}, want: &models.Attribute{Key: "key", Value: int64(987)}},
		{name: "byte type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "byte"}, want: &models.Attribute{Key: "key", Value: int64(1)}},
		{name: "byte type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "byte"}, want: &models.Attribute{Key: "key", Value: int64(0)}},
		{name: "boolean type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "boolean"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "boolean type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "boolean"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "boolean type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "boolean"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "boolean type - int64", in: &attributes.Attribute{Key: "key", Value: int64(1), Type: "boolean"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "boolean type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "boolean"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "boolean type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "boolean"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "boolean type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "boolean"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "boolean type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "boolean"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: "bool"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: "bool"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "bool type - int", in: &attributes.Attribute{Key: "key", Value: 1, Type: "bool"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "bool type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("0"), Type: "bool"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "bool type - double", in: &attributes.Attribute{Key: "key", Value: 1.1, Type: "bool"}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "bool type - bool", in: &attributes.Attribute{Key: "key", Value: false, Type: "bool"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "bool type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "bool"}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "double type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(0)}},
		{name: "double type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "double"}, want: &models.Attribute{Key: "key", Value: 123.45678}},
		{name: "double type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(987)}},
		{name: "double type - int64", in: &attributes.Attribute{Key: "key", Value: int64(987), Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(987)}},
		{name: "double type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(777)}},
		{name: "double type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "double"}, want: &models.Attribute{Key: "key", Value: 987.432}},
		{name: "double type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(1)}},
		{name: "double type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "double"}, want: &models.Attribute{Key: "key", Value: float64(0)}},
		{name: "float type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(0)}},
		{name: "float type - string good", in: &attributes.Attribute{Key: "key", Value: "123.45678", Type: "float"}, want: &models.Attribute{Key: "key", Value: 123.45678}},
		{name: "float type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(987)}},
		{name: "float type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(777)}},
		{name: "float type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(987.432)}},
		{name: "float type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(1)}},
		{name: "float type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "float"}, want: &models.Attribute{Key: "key", Value: float64(0)}},
		{name: "date type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "date type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21", Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Date(2024, 07, 21, 0, 0, 0, 0, time.UTC)}},
		{name: "date type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "date type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "date type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "date type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "date type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "date"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - string good", in: &attributes.Attribute{Key: "key", Value: "17:31:58", Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Date(0, 1, 1, 17, 31, 58, 0, time.UTC)}},
		{name: "time type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "time type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "time"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC)}},
		{name: "dateTime type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "dateTime type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "datetime"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - string bad", in: &attributes.Attribute{Key: "key", Value: "oops", Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - string good", in: &attributes.Attribute{Key: "key", Value: "2024-07-21T07:59:11Z", Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Date(2024, 07, 21, 7, 59, 11, 0, time.UTC)}},
		{name: "timestamp type - int", in: &attributes.Attribute{Key: "key", Value: 987, Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - json number", in: &attributes.Attribute{Key: "key", Value: json.Number("777"), Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - double", in: &attributes.Attribute{Key: "key", Value: 987.432, Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - bool", in: &attributes.Attribute{Key: "key", Value: true, Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "timestamp type - unsupported", in: &attributes.Attribute{Key: "key", Value: []float32{1, 2, 3}, Type: "timestamp"}, want: &models.Attribute{Key: "key", Value: time.Time{}}},
		{name: "xsd:any - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixAny}, want: &models.Attribute{Key: "key", Value: "value"}},
		{name: "xsd:any - int", in: &attributes.Attribute{Key: "key", Value: 654321, Type: xsd.PrefixAny}, want: &models.Attribute{Key: "key", Value: 654321}},
		{name: "xsd:any - double", in: &attributes.Attribute{Key: "key", Value: 555.888, Type: xsd.PrefixAny}, want: &models.Attribute{Key: "key", Value: 555.888}},
		{name: "xsd:any - json.number", in: &attributes.Attribute{Key: "key", Value: json.Number("765"), Type: xsd.PrefixAny}, want: &models.Attribute{Key: "key", Value: json.Number("765")}},
		{name: "xsd:anyURI - string", in: &attributes.Attribute{Key: "key", Value: "https://localhost:443", Type: xsd.PrefixString}, want: &models.Attribute{Key: "key", Value: "https://localhost:443"}},
		{name: "xsd:bool type - string bad", in: &attributes.Attribute{Key: "key", Value: "yeah", Type: xsd.PrefixBoolean}, want: &models.Attribute{Key: "key", Value: false}},
		{name: "xsd:bool type - string good", in: &attributes.Attribute{Key: "key", Value: "true", Type: xsd.PrefixBoolean}, want: &models.Attribute{Key: "key", Value: true}},
		{name: "xsd:string - string", in: &attributes.Attribute{Key: "key", Value: "value", Type: xsd.PrefixString}, want: &models.Attribute{Key: "key", Value: "value"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := AttributeFromOAS(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestConvertToOAS(t *testing.T) {
	testCases := []struct {
		name string
		in   *models.Attribute
		want *attributes.Attribute
	}{
		{name: "string", in: &models.Attribute{Key: "key", Value: "value"}, want: &attributes.Attribute{Key: "key", Value: "value", Type: "xsd:string"}},
		{name: "int64", in: &models.Attribute{Key: "key", Value: int64(9876543)}, want: &attributes.Attribute{Key: "key", Value: int64(9876543), Type: "xsd:long"}},
		{name: "double", in: &models.Attribute{Key: "key", Value: 123.5678}, want: &attributes.Attribute{Key: "key", Value: 123.5678, Type: "xsd:double"}},
		{name: "bool", in: &models.Attribute{Key: "key", Value: true}, want: &attributes.Attribute{Key: "key", Value: true, Type: "xsd:boolean"}},
		{name: "time.Time", in: &models.Attribute{Key: "key", Value: time.Date(2024, 5, 29, 12, 0, 0, 0, time.UTC)}, want: &attributes.Attribute{Key: "key", Value: "2024-05-29T12:00:00Z", Type: "xsd:dateTime"}},
		{name: "unsupported", in: &models.Attribute{Key: "key", Value: []float32{1, 2, 3}}, want: &attributes.Attribute{Key: "key", Value: "[1 2 3]", Type: "xsd:string"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := AttributeToOAS(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
