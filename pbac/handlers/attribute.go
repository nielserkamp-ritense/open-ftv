// Package handlers contains support for integration with HTTP services.
package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

// AttributeFromOAS converts an OAS attribute model to the internal model.
func AttributeFromOAS(in *attributes.Attribute) models.Attribute {
	value := in.Value

	if in.Type != "" {
		// we have a preference for xsd types.
		if f := conversions2[in.Type]; f != nil {
			value = f(in.Value)
		} else if f = conversions1[strings.ToLower(in.Type)]; f != nil {
			value = f(in.Value)
		}
	}

	return models.NewAttributeWithType(in.Key, value, in.Type)
}

// AttributeToOAS converts an internal attribute model to the OAS model.
func AttributeToOAS(in models.Attribute) *attributes.Attribute {
	a := &attributes.Attribute{Key: in.Key(), Value: in.Value(), Type: in.Type()}

	if a.Type == "" {
		switch t := a.Value.(type) {
		case string:
			a.Type = xsd.PrefixString
		case int64:
			a.Type = xsd.PrefixLong
		case float64:
			a.Type = xsd.PrefixDouble
		case bool:
			a.Type = xsd.PrefixBoolean
		case time.Time:
			a.Value, a.Type = t.Format(time.RFC3339Nano), xsd.PrefixDateTime
		default:
			a.Value, a.Type = fmt.Sprintf("%v", t), xsd.PrefixString
		}
	}

	return a
}

type converter func(in any) any

var conversions1 = map[string]converter{
	"bool":      anyToBool,
	"boolean":   anyToBool,
	"byte":      anyToInt64,
	"date":      anyToDate,
	"datetime":  anyToTimestamp,
	"double":    anyToFloat,
	"float":     anyToFloat,
	"int":       anyToInt64,
	"integer":   anyToInt64,
	"long":      anyToInt64,
	"short":     anyToInt64,
	"string":    anyToString,
	"time":      anyToTime,
	"timestamp": anyToTimestamp,
}

var conversions2 = map[string]converter{
	xsd.PrefixAny:              anyToAny,
	xsd.PrefixAnyURI:           anyToString,
	xsd.PrefixBoolean:          anyToBool,
	xsd.PrefixByte:             anyToInt64,
	xsd.PrefixDate:             anyToDate,
	xsd.PrefixDateTime:         anyToTimestamp,
	xsd.PrefixDay:              anyToInt64,
	xsd.PrefixDecimal:          anyToFloat,
	xsd.PrefixDouble:           anyToFloat,
	xsd.PrefixFloat:            anyToFloat,
	xsd.PrefixInt:              anyToInt64,
	xsd.PrefixInteger:          anyToInt64,
	xsd.PrefixLanguage:         anyToString,
	xsd.PrefixLong:             anyToInt64,
	xsd.PrefixMonth:            anyToInt64,
	xsd.PrefixNeg:              anyToInt64,
	xsd.PrefixNonNeg:           anyToInt64,
	xsd.PrefixNonPos:           anyToInt64,
	xsd.PrefixNormalizedString: anyToString,
	xsd.PrefixPos:              anyToInt64,
	xsd.PrefixShort:            anyToInt64,
	xsd.PrefixSimple:           anyToAny,
	xsd.PrefixString:           anyToString,
	xsd.PrefixTime:             anyToTime,
	xsd.PrefixToken:            anyToString,
	xsd.PrefixUByte:            anyToInt64,
	xsd.PrefixUInt:             anyToInt64,
	xsd.PrefixULong:            anyToInt64,
	xsd.PrefixUShort:           anyToInt64,
	xsd.PrefixYear:             anyToInt64,
}

func anyToAny(in any) any {
	return in
}

func anyToString(in any) any {
	switch t := in.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", in)
	}
}

func anyToInt64(in any) any {
	var i int64

	switch t := in.(type) {
	case json.Number:
		i, _ = t.Int64()
	case int:
		return int64(t)
	case int64:
		return t
	case float64:
		return int64(t)
	case bool:
		if t {
			i = 1
		}
	case string:
		i, _ = strconv.ParseInt(t, 0, 64)
	default:
		i, _ = strconv.ParseInt(fmt.Sprintf("%v", t), 0, 64)
	}

	return i
}

func anyToFloat(in any) any {
	var i float64

	switch t := in.(type) {
	case json.Number:
		i, _ = t.Float64()
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case bool:
		if t {
			i = 1
		}
	case string:
		i, _ = strconv.ParseFloat(t, 64)
	default:
		i, _ = strconv.ParseFloat(fmt.Sprintf("%v", t), 64)
	}

	return i
}

func anyToBool(in any) any {
	switch t := in.(type) {
	case string:
		return isTrue(t)
	case json.Number:
		return isTrue(t.String())
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case bool:
		return t
	default:
		return isTrue(fmt.Sprintf("%v", in))
	}
}

func isTrue(in string) bool {
	return strings.EqualFold(in, "true") || in == "1"
}

func anyToDate(in any) any {
	switch t := in.(type) {
	case string:
		return stringToDate(t)
	default:
		return stringToDate(fmt.Sprintf("%v", in))
	}
}

func stringToDate(in string) time.Time {
	if d, err := time.Parse("2006-01-02", in); err == nil {
		return d
	}
	return time.Time{}
}

func anyToTime(in any) any {
	switch t := in.(type) {
	case string:
		return stringToTime(t)
	default:
		return stringToTime(fmt.Sprintf("%v", in))
	}
}

func stringToTime(in string) time.Time {
	if t, err := time.Parse("15:04:05", in); err == nil {
		return t
	}
	return time.Time{}
}

func anyToTimestamp(in any) any {
	switch t := in.(type) {
	case string:
		return stringToTimestamp(t)
	default:
		return stringToTimestamp(fmt.Sprintf("%v", in))
	}
}

func stringToTimestamp(in string) time.Time {
	if ts, err := time.Parse(time.RFC3339Nano, in); err == nil {
		return ts
	}
	return time.Time{}
}
