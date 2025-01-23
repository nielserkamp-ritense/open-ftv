package xsd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLiteral(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a literal", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#label"},
		{name: "XSD any", t: "http://www.w3.org/2001/XMLSchema#anyType", want: true},
		{name: "XSD anyURI", t: "http://www.w3.org/2001/XMLSchema#anyURI", want: true},
		{name: "XSD boolean", t: "http://www.w3.org/2001/XMLSchema#boolean", want: true},
		{name: "XSD byte", t: "http://www.w3.org/2001/XMLSchema#byte", want: true},
		{name: "XSD date", t: "http://www.w3.org/2001/XMLSchema#date", want: true},
		{name: "XSD dateTime", t: "http://www.w3.org/2001/XMLSchema#dateTime", want: true},
		{name: "XSD day", t: "http://www.w3.org/2001/XMLSchema#gDay", want: true},
		{name: "XSD decimal", t: "http://www.w3.org/2001/XMLSchema#decimal", want: true},
		{name: "XSD double", t: "http://www.w3.org/2001/XMLSchema#double", want: true},
		{name: "XSD duration", t: "http://www.w3.org/2001/XMLSchema#duration", want: true},
		{name: "XSD float", t: "http://www.w3.org/2001/XMLSchema#float", want: true},
		{name: "XSD int", t: "http://www.w3.org/2001/XMLSchema#int", want: true},
		{name: "XSD integer", t: "http://www.w3.org/2001/XMLSchema#integer", want: true},
		{name: "XSD language", t: "http://www.w3.org/2001/XMLSchema#language", want: true},
		{name: "XSD long", t: "http://www.w3.org/2001/XMLSchema#long", want: true},
		{name: "XSD month", t: "http://www.w3.org/2001/XMLSchema#gMonth", want: true},
		{name: "XSD monthDay", t: "http://www.w3.org/2001/XMLSchema#gMonthDay", want: true},
		{name: "XSD neg", t: "http://www.w3.org/2001/XMLSchema#negativeInteger", want: true},
		{name: "XSD nonNeg", t: "http://www.w3.org/2001/XMLSchema#nonNegativeInteger", want: true},
		{name: "XSD nonPos", t: "http://www.w3.org/2001/XMLSchema#nonPositiveInteger", want: true},
		{name: "XSD pos", t: "http://www.w3.org/2001/XMLSchema#positiveInteger", want: true},
		{name: "XSD normalizedString", t: "http://www.w3.org/2001/XMLSchema#normalizedString", want: true},
		{name: "XSD short", t: "http://www.w3.org/2001/XMLSchema#short", want: true},
		{name: "XSD simple", t: "http://www.w3.org/2001/XMLSchema#simpleType", want: true},
		{name: "XSD string", t: "http://www.w3.org/2001/XMLSchema#string", want: true},
		{name: "XSD time", t: "http://www.w3.org/2001/XMLSchema#time", want: true},
		{name: "XSD token", t: "http://www.w3.org/2001/XMLSchema#token", want: true},
		{name: "XSD ubyte", t: "http://www.w3.org/2001/XMLSchema#unsignedByte", want: true},
		{name: "XSD uint", t: "http://www.w3.org/2001/XMLSchema#unsignedInt", want: true},
		{name: "XSD ulong", t: "http://www.w3.org/2001/XMLSchema#unsignedLong", want: true},
		{name: "XSD ushort", t: "http://www.w3.org/2001/XMLSchema#unsignedShort", want: true},
		{name: "XSD year", t: "http://www.w3.org/2001/XMLSchema#gYear", want: true},
		{name: "XSD yearMonth", t: "http://www.w3.org/2001/XMLSchema#gYearMonth", want: true},
		{name: "XSD any (2)", t: "xsd:anyType", want: true},
		{name: "XSD anyURI (2)", t: "xsd:anyURI", want: true},
		{name: "XSD boolean (2)", t: "xsd:boolean", want: true},
		{name: "XSD byte (2)", t: "xsd:byte", want: true},
		{name: "XSD date (2)", t: "xsd:date", want: true},
		{name: "XSD dateTime (2)", t: "xsd:dateTime", want: true},
		{name: "XSD day (2)", t: "xsd:gDay", want: true},
		{name: "XSD decimal (2)", t: "xsd:decimal", want: true},
		{name: "XSD double (2)", t: "xsd:double", want: true},
		{name: "XSD duration (2)", t: "xsd:duration", want: true},
		{name: "XSD float (2)", t: "xsd:float", want: true},
		{name: "XSD int (2)", t: "xsd:int", want: true},
		{name: "XSD integer (2)", t: "xsd:integer", want: true},
		{name: "XSD language (2)", t: "xsd:language", want: true},
		{name: "XSD long (2)", t: "xsd:long", want: true},
		{name: "XSD month (2)", t: "xsd:gMonth", want: true},
		{name: "XSD monthDay (2)", t: "xsd:gMonthDay", want: true},
		{name: "XSD neg (2)", t: "xsd:negativeInteger", want: true},
		{name: "XSD nonNeg (2)", t: "xsd:nonNegativeInteger", want: true},
		{name: "XSD nonPos (2)", t: "xsd:nonPositiveInteger", want: true},
		{name: "XSD pos (2)", t: "xsd:positiveInteger", want: true},
		{name: "XSD normalizedString (2)", t: "xsd:normalizedString", want: true},
		{name: "XSD short (2)", t: "xsd:short", want: true},
		{name: "XSD simple (2)", t: "xsd:simpleType", want: true},
		{name: "XSD string (2)", t: "xsd:string", want: true},
		{name: "XSD time (2)", t: "xsd:time", want: true},
		{name: "XSD token (2)", t: "xsd:token", want: true},
		{name: "XSD ubyte (2)", t: "xsd:unsignedByte", want: true},
		{name: "XSD uint (2)", t: "xsd:unsignedInt", want: true},
		{name: "XSD ulong (2)", t: "xsd:unsignedLong", want: true},
		{name: "XSD ushort (2)", t: "xsd:unsignedShort", want: true},
		{name: "XSD year (2)", t: "xsd:gYear", want: true},
		{name: "XSD yearMonth (2)", t: "xsd:gYearMonth", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsLiteral(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}
