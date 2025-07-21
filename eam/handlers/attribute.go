// Package handlers contains support for integration with HTTP services.
package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

// AttributeFromOAS converts an OAS attribute model to the internal model.
func AttributeFromOAS(in *attributes.Attribute) *models.Attribute {
	value := in.Value

	if in.Type != "" {
		// we have a preference for xsd types.
		if f := conversions2[in.Type]; f != nil {
			value = f(in.Value)
		} else if f = conversions1[strings.ToLower(in.Type)]; f != nil {
			value = f(in.Value)
		}
	}

	return models.NewOriginalAttribute(in.Key, value, in.Value, in.Type)
}

// AttributeToOAS converts an internal attribute model to the OAS model.
func AttributeToOAS(in *models.Attribute) *attributes.Attribute {
	a := &attributes.Attribute{Key: in.Key(), Value: in.Original(), Type: in.Type()}

	if a.Value == nil {
		a.Value = in.Value()
	}

	if a.Type == "" {
		switch t := a.Value.(type) {
		case string:
			a.Type = "string"
		case int64:
			a.Type = "long"
		case float64:
			a.Type = "double"
		case json.Number:
			a.Type = "long"
		case bool:
			a.Type = "bool"
		case time.Time:
			a.Value, a.Type = t.Format(time.RFC3339Nano), "datetime"
		case time.Duration:
			a.Value, a.Type = t.String(), "duration"
		case *models.AttributeSet:
			a.Value = models.MapFromAttributes(t)
		// next types we leave the value and type as-is.
		case []any, map[string]any:
		case int, int8, int16, int32:
		case uint, uint8, uint16, uint32, uint64:
		default:
			// everything else: convert to string.
			a.Value, a.Type = fmt.Sprintf("%v", t), "string"
		}
	} else {
		switch a.Type {
		case "date", xsd.PrefixDate:
			if d, ok := anyToDate(a.Value).(time.Time); ok {
				a.Value = d.Format("2006-01-02")
			}

		case "time", xsd.PrefixTime:
			if d, ok := anyToTime(a.Value).(time.Time); ok {
				a.Value = d.Format("15:04:05")
			}

		case "datetime", "timestamp", xsd.PrefixDateTime:
			if d, ok := anyToTimestamp(a.Value).(time.Time); ok {
				a.Value = d.Format(time.RFC3339Nano)
			}

		case "duration":
			if d, ok := anyToDuration(a.Value).(time.Duration); ok {
				a.Value = d.String()
			}

		case xsd.PrefixDuration:
			if d, ok := anyToDuration(a.Value).(time.Duration); ok {
				s := d.String()
				if strings.HasSuffix(s, "ms") || strings.HasSuffix(s, "µs") || strings.HasSuffix(s, "ns") {
					d += time.Second
					s = d.String()
					s = "0" + s[1:]
				}
				a.Value = "PT" + strings.ToUpper(s)
			}
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
	"duration":  anyToDuration,
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
	xsd.PrefixAny:        anyToAny,
	xsd.PrefixAnyURI:     anyToString,
	xsd.PrefixBoolean:    anyToBool,
	xsd.PrefixByte:       anyToInt64,
	xsd.PrefixDate:       anyToDate,
	xsd.PrefixDateTime:   anyToTimestamp,
	xsd.PrefixDay:        anyToInt64,
	xsd.PrefixDecimal:    anyToFloat,
	xsd.PrefixDouble:     anyToFloat,
	xsd.PrefixDuration:   anyToDuration,
	xsd.PrefixFloat:      anyToFloat,
	xsd.PrefixInt:        anyToInt64,
	xsd.PrefixInteger:    anyToInt64,
	xsd.PrefixLanguage:   anyToString,
	xsd.PrefixLong:       anyToInt64,
	xsd.PrefixMonth:      anyToInt64,
	xsd.PrefixMonthDay:   anyToString,
	xsd.PrefixNeg:        anyToInt64,
	xsd.PrefixNonNeg:     anyToInt64,
	xsd.PrefixNonPos:     anyToInt64,
	xsd.PrefixNormalized: anyToString,
	xsd.PrefixPos:        anyToInt64,
	xsd.PrefixShort:      anyToInt64,
	xsd.PrefixSimple:     anyToAny,
	xsd.PrefixString:     anyToString,
	xsd.PrefixTime:       anyToTime,
	xsd.PrefixToken:      anyToString,
	xsd.PrefixUByte:      anyToInt64,
	xsd.PrefixUInt:       anyToInt64,
	xsd.PrefixULong:      anyToInt64,
	xsd.PrefixUShort:     anyToInt64,
	xsd.PrefixYear:       anyToInt64,
	xsd.PrefixYearMonth:  anyToString,
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
		return strconv.FormatFloat(t, 'g', -1, 64)
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
		i = int64(t)
	case int64:
		i = t
	case float64:
		i = int64(t)
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
		i = t
	case int:
		i = float64(t)
	case int64:
		i = float64(t)
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
	case bool:
		return t
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
	default:
		return isTrue(fmt.Sprintf("%v", in))
	}
}

func isTrue(in string) bool {
	return strings.EqualFold(in, "true") || in == "1"
}

func anyToDate(in any) any {
	switch t := in.(type) {
	case time.Time:
		return t
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
	case time.Time:
		return t
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
	case time.Time:
		return t
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

func anyToDuration(in any) any {
	switch t := in.(type) {
	case time.Duration:
		return t
	case string:
		return stringToDuration(t)
	default:
		return stringToDuration(fmt.Sprintf("%v", in))
	}
}

func stringToDuration(in string) time.Duration {
	if d, err := time.ParseDuration(in); err == nil {
		return d
	}

	if a, err := xsd.FromString(in, xsd.PrefixDuration); err == nil {
		if d, ok := a.(time.Duration); ok {
			return d
		}
	}

	return 0
}
