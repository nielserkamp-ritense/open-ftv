package rdf

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// ConvertXSD converts an RDF literal into a Golang variable of the appropriate type.
//
// This function supports most of the common XSD data-types.
func ConvertXSD(data, t string) (any, error) {
	switch t {
	case XSDString, XSDNormalizedString, XSDToken, XSDLanguage, XSDAny, XSDSimple:
		return data, nil
	case XSDBoolean:
		return data == "true" || data == "1", nil
	case XSDFloat:
		return strconv.ParseFloat(data, 32)
	case XSDDouble:
		return strconv.ParseFloat(data, 64)
	case XSDDecimal:
		return convertDecimal(data)
	case XSDInteger, XSDLong, XSDNonPos, XSDNeg:
		return strconv.ParseInt(data, 10, 64)
	case XSDNonNeg, XSDPos, XSDULong:
		return strconv.ParseUint(data, 10, 64)
	case XSDInt:
		return strconv.ParseInt(data, 10, 32)
	case XSDShort:
		return strconv.ParseInt(data, 10, 16)
	case XSDByte:
		return strconv.ParseInt(data, 10, 8)
	case XSDUInt, XSDYear:
		return strconv.ParseUint(data, 10, 32)
	case XSDUShort:
		return strconv.ParseUint(data, 10, 16)
	case XSDUByte, XSDDay, XSDMonth:
		return strconv.ParseUint(data, 10, 8)
	case XSDDuration:
		return convertDuration(data)
	case XSDDateTime:
		return time.ParseInLocation(time.RFC3339Nano, data, time.Local)
	case XSDAnyURI:
		return convertURI(data)

	case XSDTime:
		// TODO: we may need a special time type.
		return time.ParseInLocation("15:04:05.999999999", data, time.Local)
	case XSDDate:
		// TODO: we may need a special date type.
		return time.ParseInLocation("2006-01-02", data, time.Local)
	case XSDYearMonth:
		// TODO: we may need a special year-month type.
		return time.ParseInLocation("2006-01", data, time.Local)
	case XSDMonthDay:
		// TODO: we may need a special month-day type.
		return time.ParseInLocation("01-02", data, time.Local)

	default:
		return nil, fmt.Errorf("rdf: invalid data type: <%s>", t)
	}
}

func convertDecimal(data string) (any, error) {
	if i, err := strconv.ParseInt(data, 10, 64); err == nil {
		return i, nil
	}
	return strconv.ParseFloat(data, 64)
}

func convertDuration(data string) (any, error) {
	if !rx1.MatchString(data) {
		return 0, fmt.Errorf("rdf: invalid duration: <%s>", data)
	}

	capture := rx1.FindStringSubmatch(data)

	if capture[2] != "" || capture[3] != "" || capture[4] != "" {
		return data, nil
	}

	var out time.Duration

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
		out += time.Duration(f) * time.Second
	}

	return out, nil
}

func convertURI(data string) (any, error) {
	u, err := url.Parse(data)
	if err != nil {
		return nil, err
	}
	return u, nil
}

var rx1 = regexp.MustCompile(`^(-?)P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:([\d.]+)S)?)?$`)
