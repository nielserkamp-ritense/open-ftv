package xsd

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// FromString converts an RDF literal into a variable of the appropriate Golang data-type.
//
// This function supports most of the common XSD data-types.
func FromString(data, t string) (any, error) {
	switch t {
	case URIString, URINormalized, URIToken, URILanguage, URIAny, URISimple, PrefixString, PrefixNormalized, PrefixToken, PrefixLanguage, PrefixAny, PrefixSimple:
		return data, nil
	case URIBoolean, PrefixBoolean:
		return data == "true" || data == "1", nil
	case URIFloat, PrefixFloat:
		return strconv.ParseFloat(data, 32)
	case URIDouble, PrefixDouble:
		return strconv.ParseFloat(data, 64)
	case URIDecimal, PrefixDecimal:
		return fromDecimalString(data)
	case URIInteger, URILong, URINonPos, URINeg, PrefixInteger, PrefixLong, PrefixNonPos, PrefixNeg:
		return strconv.ParseInt(data, 10, 64)
	case URINonNeg, URIPos, URIULong, PrefixNonNeg, PrefixPos, PrefixULong:
		return strconv.ParseUint(data, 10, 64)
	case URIInt, PrefixInt:
		return strconv.ParseInt(data, 10, 32)
	case URIShort, PrefixShort:
		return strconv.ParseInt(data, 10, 16)
	case URIByte, PrefixByte:
		return strconv.ParseInt(data, 10, 8)
	case URIUInt, URIYear, PrefixUInt, PrefixYear:
		return strconv.ParseUint(data, 10, 32)
	case URIUShort, PrefixUShort:
		return strconv.ParseUint(data, 10, 16)
	case URIUByte, URIDay, URIMonth, PrefixUByte, PrefixDay, PrefixMonth:
		return strconv.ParseUint(data, 10, 8)
	case URIDuration, PrefixDuration:
		return fromDurationString(data)
	case URIDateTime, PrefixDateTime:
		return time.ParseInLocation(time.RFC3339Nano, data, time.Local)
	case URIAnyURI, PrefixAnyURI:
		return fromURIString(data)

	case URITime, PrefixTime:
		// TODO: we may need a special time type.
		return time.ParseInLocation("15:04:05.999999999", data, time.Local)
	case URIDate, PrefixDate:
		// TODO: we may need a special date type.
		return time.ParseInLocation("2006-01-02", data, time.Local)
	case URIYearMonth, PrefixYearMonth:
		// TODO: we may need a special year-month type.
		return time.ParseInLocation("2006-01", data, time.Local)
	case URIMonthDay, PrefixMonthDay:
		// TODO: we may need a special month-day type.
		return time.ParseInLocation("01-02", data, time.Local)

	default:
		return nil, fmt.Errorf("rdf: invalid data type: <%s>", t)
	}
}

func fromDecimalString(data string) (any, error) {
	if i, err := strconv.ParseInt(data, 10, 64); err == nil {
		return i, nil
	}
	return strconv.ParseFloat(data, 64)
}

func fromDurationString(data string) (any, error) {
	var out time.Duration

	if !rx1.MatchString(data) {
		return out, fmt.Errorf("rdf: invalid duration: <%s>", data)
	}

	capture := rx1.FindStringSubmatch(data)

	if capture[2] != "" || capture[3] != "" || capture[4] != "" {
		return data, nil
	}

	if s := capture[5]; s != "" {
		i, _ := strconv.ParseInt(capture[5], 10, 64)
		out += time.Duration(i) * time.Hour
	}

	if s := capture[6]; s != "" {
		i, _ := strconv.ParseInt(s, 10, 64)
		out += time.Duration(i) * time.Minute
	}

	if s := capture[7]; s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("rdf: invalid duration: <%s>", data)
		}
		out += time.Duration(f * float64(time.Second))
	}

	if capture[1] == "-" {
		out *= -1
	}

	return out, nil
}

func fromURIString(data string) (any, error) {
	u, err := url.Parse(data)
	if err != nil {
		return nil, err
	}
	return u, nil
}

var rx1 = regexp.MustCompile(`^(-?)P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:([\d.]+)S)?)?$`)
