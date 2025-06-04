package csv

import (
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// Encoder represents the generic interface for encoding a value to CSV format.
type Encoder interface {
	Encode() any
}

func encode(def *schema.Field, v any) any {
	if def.IsArray {
		return encodeValues(def, v)
	}
	return encodeValue(def, v)
}

func encodeValues(def *schema.Field, v any) any {
	switch def.Type {
	case enums.StringType, enums.URLType, enums.EmailType, enums.PhoneNrType, enums.IPAddressType:
		return encodeStrings(def, v)
	case enums.IntegerType:
		return encodeStrings(def, convert.AnyToInts(v))
	case enums.UnsignedIntegerType:
		return encodeStrings(def, convert.AnyToUints(v))
	case enums.FloatType:
		return encodeStrings(def, convert.AnyToFloats(v))
	case enums.BooleanType:
		return encodeStrings(def, convert.AnyToBools(v))
	case enums.DateType, enums.TimeType, enums.DateTimeType:
		// TODO: encode slice of dates, times or timestamps.
	case enums.ObjectType, enums.AnyType:
		return encodeObjects(def, v)
	}

	return encodeStrings(def, v)
}

func encodeValue(def *schema.Field, v any) any {
	switch def.Type {
	case enums.StringType, enums.URLType, enums.EmailType, enums.PhoneNrType, enums.IPAddressType:
		return convert.AnyToString(v)
	case enums.IntegerType:
		return convert.AnyToInt64(v)
	case enums.UnsignedIntegerType:
		return convert.AnyToUint64(v)
	case enums.FloatType:
		return convert.AnyToFloat64(v)
	case enums.BooleanType:
		return convert.AnyToBool(v)
	case enums.DateType, enums.TimeType, enums.DateTimeType:
		return convert.AnyToDateTime(v)
	case enums.ObjectType:
		return encodeObject(def, v)
	default:
		return encodeAny(v)
	}
}

func encodeStrings(_ *schema.Field, v any) any {
	list := convert.AnyToStrings(v)

	// TODO: perhaps a different separator or encoding?
	return strings.Join(list, ",")
}

func encodeObjects(def *schema.Field, v any) any {
	switch t := v.(type) {
	case []any:
		return encodeAnySlice(def, t)
	case []map[string]any:
		return encodeMaps(def, t)
	case [][]any:
		return encodeSlices(def, t)
	default:
		return encodeAny(v)
	}
}

func encodeAnySlice(def *schema.Field, v []any) any {
	out := make([]any, len(v))
	for i := range v {
		out[i] = encodeValue(def, v[i])
	}
	return out
}

func encodeMaps(def *schema.Field, v []map[string]any) any {
	out := make([]map[string]any, len(v))

	for i := range v {
		rec := v[i]
		out[i] = make(map[string]any, len(def.Fields))

		for _, field := range def.Fields {
			v2 := rec[field.ID]
			out[i][field.ID] = encode(field, v2)
		}
	}

	return out
}

func encodeSlices(def *schema.Field, v [][]any) any {
	out := make([]map[string]any, len(v))

	for i := range v {
		rec := v[i]
		out[i] = make(map[string]any, len(def.Fields))

		for j, field := range def.Fields {
			v2 := rec[j]
			out[i][field.ID] = encode(field, v2)
		}
	}

	return out
}

func encodeObject(def *schema.Field, v any) any {
	switch t := v.(type) {
	case map[string]any:
		return encodeMap(def, t)
	case []any:
		return encodeSlice(def, t)
	default:
		return encodeAny(v)
	}
}

func encodeMap(def *schema.Field, v map[string]any) any {
	out := make(map[string]any, len(def.Fields))
	for _, field := range def.Fields {
		v2 := v[field.ID]
		out[field.ID] = encode(field, v2)
	}
	return out
}

func encodeSlice(def *schema.Field, v []any) any {
	out := make(map[string]any, len(def.Fields))
	for i, field := range def.Fields {
		v2 := v[i]
		out[field.ID] = encode(field, v2)
	}
	return out
}

func encodeAny(v any) any {
	if e, ok := v.(Encoder); ok {
		return e.Encode()
	}
	return convert.AnyToString(v)
}

func encodeBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func encodeTime(def *schema.Field, t time.Time) string {
	var format string

	switch {
	case t.IsZero():
		return ""
	case def.Format == "brp-date":
		format = "20060102"
	case def.Format != "":
		format = def.Format
	case t.Year() <= 1800:
		format = "15:04:05"
	case t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() < 1000000:
		format = "2006-01-02"
	case t.Nanosecond() < 1000000:
		format = "2006-01-02 15:04:05"
	default:
		format = "2006-01-02 15:04:05.999"
	}

	return t.Format(format)
}
