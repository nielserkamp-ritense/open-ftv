package rdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertLiteral(t *testing.T) {
	html := "<html><head><title>Hello World</title></head><body>nothing to see here</body></html>"
	xml := "<xml><title>Hello World</title></xml>"
	j1 := `{"title":"Hello World"}`
	j2 := map[string]any{"title": "Hello World"}

	testCases := []struct {
		name    string
		data    string
		t       string
		want    any
		wantErr bool
	}{
		{name: "no mime-type", data: "abc", wantErr: true},
		{name: "bad mime-type", data: "abc", t: "abc", wantErr: true},
		{name: "string", data: "abc", t: "http://www.w3.org/2001/XMLSchema#string", want: "abc"},
		{name: "html", data: html, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML", want: html},
		{name: "xml", data: xml, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral", want: xml},
		{name: "json", data: j1, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#JSON", want: j2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertLiteral(tc.data, tc.t)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestIsClass(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a class", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDF property", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Property", want: true},
		{name: "RDFS class", t: "http://www.w3.org/2000/01/rdf-schema#Class", want: true},
		{name: "RDFS resource", t: "http://www.w3.org/2000/01/rdf-schema#Resource", want: true},
		{name: "OWL class", t: "http://www.w3.org/2002/07/owl#Class", want: true},
		{name: "OWL thing", t: "http://www.w3.org/2002/07/owl#Thing", want: true},
		{name: "SKOS concept", t: "http://www.w3.org/2004/02/skos/core#Concept", want: true},
		{name: "FTV attribute", t: "https://ftv.nl/rdf/ftv#Attribute", want: true},
		{name: "FTV entity-attribute", t: "https://ftv.nl/rdf/ftv#entityAttribute", want: true},
		{name: "FTV entity", t: "https://ftv.nl/rdf/ftv#Entity", want: true},
		{name: "FTV relation", t: "https://ftv.nl/rdf/ftv#Relation", want: true},
		{name: "FTV principal", t: "https://ftv.nl/rdf/ftv#principal", want: true},
		{name: "FTV resource", t: "https://ftv.nl/rdf/ftv#resource", want: true},
		{name: "FTV context", t: "https://ftv.nl/rdf/ftv#context", want: true},
		{name: "FTV subject", t: "https://ftv.nl/rdf/ftv#subject", want: true},
		{name: "FTV object", t: "https://ftv.nl/rdf/ftv#object", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsClass(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsLabel(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a label", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDFS label", t: "http://www.w3.org/2000/01/rdf-schema#label", want: true},
		{name: "SKOS prefLabel", t: "http://www.w3.org/2004/02/skos/core#prefLabel", want: true},
		{name: "SKOS altLabel", t: "http://www.w3.org/2004/02/skos/core#altLabel", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsLabel(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsTitle(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a title", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "DC terms title", t: "http://purl.org/dc/terms/title", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTitle(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsList(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a list", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDF bag", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag", want: true},
		{name: "RDF seq", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq", want: true},
		{name: "RDF list", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#List", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsList(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsAttribute(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not an attribute", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV attribute", t: "https://ftv.nl/rdf/ftv#Attribute", want: true},
		{name: "FTV entity-attribute", t: "https://ftv.nl/rdf/ftv#entityAttribute", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAttribute(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsEntity(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not an entity", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV entity", t: "https://ftv.nl/rdf/ftv#Entity", want: true},
		{name: "FTV principal", t: "https://ftv.nl/rdf/ftv#principal", want: true},
		{name: "FTV resource", t: "https://ftv.nl/rdf/ftv#resource", want: true},
		{name: "FTV context", t: "https://ftv.nl/rdf/ftv#context", want: true},
		{name: "FTV subject", t: "https://ftv.nl/rdf/ftv#subject", want: true},
		{name: "FTV object", t: "https://ftv.nl/rdf/ftv#object", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEntity(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsRelation(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a relation", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV relation", t: "https://ftv.nl/rdf/ftv#Relation", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsRelation(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsValue(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a value", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#label"},
		{name: "RDF value", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#value", want: true},
		{name: "RDF property", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Property", want: true},
		{name: "FTV attribute key", t: "https://ftv.nl/rdf/ftv#attributeKey", want: true},
		{name: "FTV entity type", t: "https://ftv.nl/rdf/ftv#entityType", want: true},
		{name: "FTV entity id", t: "https://ftv.nl/rdf/ftv#entityID", want: true},
		{name: "FTV action", t: "https://ftv.nl/rdf/ftv#action", want: true},
		{name: "FTV relation type", t: "https://ftv.nl/rdf/ftv#relationType", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValue(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsLiteral(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a literal", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#label"},
		{name: "FTV attribute key", t: "https://ftv.nl/rdf/ftv#attributeKey", want: true},
		{name: "FTV entity type", t: "https://ftv.nl/rdf/ftv#entityType", want: true},
		{name: "FTV entity id", t: "https://ftv.nl/rdf/ftv#entityID", want: true},
		{name: "FTV action", t: "https://ftv.nl/rdf/ftv#action", want: true},
		{name: "FTV relation type", t: "https://ftv.nl/rdf/ftv#relationType", want: true},
		{name: "RDF html", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML", want: true},
		{name: "RDF xml", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral", want: true},
		{name: "RDF json", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#JSON", want: true},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsLiteral(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}
