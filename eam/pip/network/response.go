package network

// ResponseMapping contains the parameters used to decode a response when pulling external data.
//
// Supported encodings are JSON, YAML, TOML, RDF/Turtle and RDF/JSON-LD.
// The server should send the Content-Type header to indicate the format of the response body.
// If it doesn't, the content type will be determined by looking at the first bytes of the body.
// This may fail for a variety of reasons, so it is recommended to always send the appropriate Content-Type header.
//
// An RDF-based response body **must** adhere to the FTV ontology.
//
// The key for each map indicates the base key where objects of that type can be found in the response.
// E.g., the response is decoded into a tree of key/value pairs (e.g., an object where any value can be another object).
// A base key contains the concatenated key values to the base of an object or array of objects with that type.
// Key values are concatenated with a dot, as common in object-oriented languages; e.g. "data.network.attributes".
// Keys that contain one or more dots must be enclosed in double or single quotes; e.g. "data.'net.addresses'.attributes".
type ResponseMapping struct {
	Attributes  []*AttributesMapping `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`    // The mappings for decoding attributes.
	Entities    []*EntityMapping     `json:"entities,omitempty" yaml:"entities,omitempty" toml:"entities,omitempty"`          // The mappings for decoding entities.
	Relations   []*RelationMapping   `json:"relations,omitempty" yaml:"relations,omitempty" toml:"relations,omitempty"`       // The mappings for decoding relations.
	StatusCodes []*StatusCode        `json:"statusCodes,omitempty" yaml:"statusCodes,omitempty" toml:"statusCodes,omitempty"` // The mappings for decoding based on response status code.
}

// AttributesMapping defines the mapping for a set of attributes.
type AttributesMapping struct {
	Base string              `json:"base" yaml:"base" toml:"base"` // the unique base key for processing a set of attributes.
	Map  []*AttributeMapping `json:"map" yaml:"map" toml:"map"`    // slice of individual attributes within the set.
}

// AttributeMapping defines the mapping for a single attribute.
//
// The value from either finding the KeyField in the data, or KeyValue,
// will be prepended with KeyParents, with all parts seperated by dots.
//
// Examples:
//  1. KeyParents := nil; KeyValue := "x" ==> full_key := "x"
//  2. KeyParents := nil; KeyValue := "x.y.z" ==> full_key := "x.y.z
//  3. KeyParents := ["a", "b"]; KeyValue := "x" ==> full_key := "a.b.x"
//  4. KeyParents := ["a", "b"]; KeyValue := "x.y.z" ==> full_key := "a.b.x.y.z
//
// The same happens if a value is found in the response data using the KeyField (replacing KeyValue in the previous examples).
//
// If the resulting full_key contains dots, it is considered to be a multi-level attribute key.
// Any key at a level that does not exist in the attribute tree will be created.
// If an existing key level does not have an object as value, its value will be replaced with an object.
type AttributeMapping struct {
	KeyParents []string `json:"keyParents,omitempty" yaml:"keyParents,omitempty" toml:"keyParents,omitempty"` // Parent key codes for the attribute.
	KeyField   string   `json:"keyField,omitempty" yaml:"keyField,omitempty" toml:"keyField,omitempty"`       // Code of the field containing the attribute key. Mutually exclusive with KeyValue.
	KeyValue   string   `json:"keyValue,omitempty" yaml:"keyValue,omitempty" toml:"keyValue,omitempty"`       // Fixed value for the attribute key. Mutually exclusive with KeyField.
	ValueField string   `json:"valueField,omitempty" yaml:"valueField,omitempty" toml:"valueField,omitempty"` // Code of the field containing the attribute value. Mutually exclusive with ValueAsIs and ValueValue.
	ValueAsIs  bool     `json:"valueAsIs,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"`    // Indicates the response object itself is the value for the attribute. Mutually exclusive with ValueField and ValueValue.
	ValueValue any      `json:"valueValue,omitempty" yaml:"valueValue,omitempty" toml:"valueValue,omitempty"` // Fixed value for the attribute value. Mutually exclusive with ValueField and ValueAsIs.
	TypeField  string   `json:"typeField,omitempty" yaml:"typeField,omitempty" toml:"typeField,omitempty"`    // Code of the field containing the type of the attribute. Mutually exclusive with TypeValue.
	TypeValue  string   `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"`    // Fixed type of the attribute. Mutually exclusive with TypeField.
}

// EntityMapping defines the mapping for an entity.
type EntityMapping struct {
	Base        string               `json:"base" yaml:"base" toml:"base"`                                                       // The unique base key for processing an entity.
	TypeField   string               `json:"typeField,omitempty" yaml:"typeField,omitempty" toml:"typeField,omitempty"`          // Code of the field containing the entity type. Mutually exclusive with TypeValue.
	TypeValue   string               `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"`          // Fixed value for the entity type. Mutually exclusive with TypeField.
	IDField     string               `json:"idField,omitempty" yaml:"idField,omitempty" toml:"idField,omitempty"`                // Code of the field containing the entity ID. Mutually exclusive with IDValue and IDFromValue.
	IDValue     string               `json:"idValue,omitempty" yaml:"idValue,omitempty" toml:"idValue,omitempty"`                // Fixed value for the entity ID. Mutually exclusive with IDField and IDFromValue.
	IDFromValue bool                 `json:"idFromValue,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"`        // Indicates the response object itself is the value for the ID. Mutually exclusive with IDValue and IDField.
	Attributes  []*AttributesMapping `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`       // Relative base(s) for processing attributes for the entity.
	ParentsCode string               `json:"parentsField,omitempty" yaml:"parentsField,omitempty" toml:"parentsField,omitempty"` // Code of the field containing the optional parent keys for the entity.
}

// RelationMapping defines the mapping for a relation.
type RelationMapping struct {
	Base               string `json:"base" yaml:"base" toml:"base"`                                                                         // The unique base key for processing a relation.
	SubjectTypeField   string `json:"subjectTypeField,omitempty" yaml:"subjectTypeField,omitempty" toml:"subjectTypeField,omitempty"`       // Code of the field containing the subject type. Mutually exclusive with SubjectTypeValue.
	SubjectTypeValue   string `json:"subjectTypeValue,omitempty" yaml:"subjectTypeValue,omitempty" toml:"subjectTypeValue,omitempty"`       // Fixed value for the subject type. Mutually exclusive with SubjectTypeField.
	SubjectIDField     string `json:"subjectIdField,omitempty" yaml:"subjectIdField,omitempty" toml:"subjectIdField,omitempty"`             // Code of the field containing the subject ID. Mutually exclusive with SubjectIDValue.
	SubjectIDValue     string `json:"subjectIdValue,omitempty" yaml:"subjectIdValue,omitempty" toml:"subjectIdValue,omitempty"`             // Fixed value for the subject ID. Mutually exclusive with SubjectIDField.
	PredicateTypeField string `json:"predicateTypeField,omitempty" yaml:"predicateTypeField,omitempty" toml:"predicateTypeField,omitempty"` // Code of the field containing the predicate type. Mutually exclusive with PredicateTypeValue.
	PredicateTypeValue string `json:"predicateTypeValue,omitempty" yaml:"predicateTypeValue,omitempty" toml:"predicateTypeValue,omitempty"` // Fixed value for the predicate type. Mutually exclusive with PredicateTypeField.
	PredicateIDField   string `json:"predicateIdField,omitempty" yaml:"predicateIdField,omitempty" toml:"predicateIdField,omitempty"`       // Code of the field containing the predicate ID. Mutually exclusive with PredicateIDValue.
	PredicateIDValue   string `json:"predicateIdValue,omitempty" yaml:"predicateIdValue,omitempty" toml:"predicateIdValue,omitempty"`       // Fixed value for the predicate ID. Mutually exclusive with PredicateIDField.
	ObjectTypeField    string `json:"objectTypeField,omitempty" yaml:"objectTypeField,omitempty" toml:"objectTypeField,omitempty"`          // Code of the field containing the object type. Mutually exclusive with ObjectTypeValue.
	ObjectTypeValue    string `json:"objectTypeValue,omitempty" yaml:"objectTypeValue,omitempty" toml:"objectTypeValue,omitempty"`          // Fixed value for the object type. Mutually exclusive with ObjectTypeField.
	ObjectIDField      string `json:"objectIdField,omitempty" yaml:"objectIdField,omitempty" toml:"objectIdField,omitempty"`                // Code of the field containing the object ID. Mutually exclusive with ObjectIDValue.
	ObjectIDValue      string `json:"objectIdValue,omitempty" yaml:"objectIdValue,omitempty" toml:"objectIdValue,omitempty"`                // Fixed value for the object ID. Mutually exclusive with ObjectIDField.
}
