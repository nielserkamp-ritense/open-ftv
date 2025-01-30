package network

// ResponseMapping contains the parameters used to decode a response when pulling external data.
//
// Supported encodings are JSON, YAML, TOML, RDF/Turtle and RDF/JSON-LD.
// The server should send the Content-Type header to indicate the format of the response body.
// If it doesn't, the content type will be determined by looking at the first bytes of the body.
// This may fail for a variety of reasons, so it is recommended to always send the appropriate Content-Type header.
//
// An RDF based response body **must** adhere to the FTV ontology.
//
// The key for each map indicates the base key where objects of that type can be found in the response.
// E.g. the response is decoded into a recursive tree of key/value pairs (an object).
// A base key contains the concatenated key values to the base of an object or array of objects of that type.
// Key values are concatenated with a dot, as common in object-oriented languages; e.g. "data.network.attributes".
// Keys that contain one or more dots, must be enclosed in double or single quotes; e.g. "data.'net.addresses'.attributes".
type ResponseMapping struct {
	Attributes []*AttributesMapping `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"` // the mappings for decoding attributes.
	Entities   []*EntityMapping     `json:"entities,omitempty" yaml:"entities,omitempty" toml:"entities,omitempty"`       // the mappings for decoding entities.
	Relations  []*RelationMapping   `json:"relations,omitempty" yaml:"relations,omitempty" toml:"relations,omitempty"`    // the mappings for decoding relations.
}

// AttributesMapping defines the mapping for a set of attributes.
type AttributesMapping struct {
	Base string              `json:"base" yaml:"base" toml:"base"` // the unique base key for processing a set of attributes.
	Map  []*AttributeMapping `json:"map" yaml:"map" toml:"map"`    // slice of individual attributes within the set.
}

// AttributeMapping defines the mapping for a single attribute.
type AttributeMapping struct {
	KeyField   string `json:"keyField,omitempty" yaml:"keyField,omitempty" toml:"keyField,omitempty"`       // code of the field containing the attribute key. Mutually exclusive with KeyValue.
	KeyValue   string `json:"keyValue,omitempty" yaml:"keyValue,omitempty" toml:"keyValue,omitempty"`       // fixed value for the attribute key. Mutually exclusive with KeyField.
	ValueField string `json:"valueField,omitempty" yaml:"valueField,omitempty" toml:"valueField,omitempty"` // code of the field containing the attribute value. Mutually exclusive with ValueAsIs.
	ValueAsIs  bool   `json:"valueAsIs,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"`    // indicates the response object itself is the value for the attribute. Mutually exclusive with ValueField.
	TypeField  string `json:"typeField,omitempty" yaml:"typeField,omitempty" toml:"typeField,omitempty"`    // code of the field containing the type of the attribute. Mutually exclusive with TypeValue.
	TypeValue  string `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"`    // fixed type of the attribute. Mutually exclusive with TypeField.
}

// EntityMapping defines the mapping for an entity.
type EntityMapping struct {
	Base        string               `json:"base" yaml:"base" toml:"base"`                                                       // the unique base key for processing an entity.
	TypeField   string               `json:"typeField,omitempty" yaml:"typeField,omitempty" toml:"typeField,omitempty"`          // code of the field containing the entity type. Mutually exclusive with TypeValue.
	TypeValue   string               `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"`          // fixed value for the entity type. Mutually exclusive with TypeField.
	IdField     string               `json:"idField,omitempty" yaml:"idField,omitempty" toml:"idField,omitempty"`                // code of the field containing the entity ID. Mutually exclusive with IdValue and IdFromValue.
	IdValue     string               `json:"idValue,omitempty" yaml:"idValue,omitempty" toml:"idValue,omitempty"`                // fixed value for the entity ID. Mutually exclusive with IdField and IdFromValue.
	IdFromValue bool                 `json:"idFromValue,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"`        // indicates the response object itself is the value for the ID. Mutually exclusive with IdValue and IdField.
	Attributes  []*AttributesMapping `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`       // relative base(s) for processing attributes for the entity.
	ParentsCode string               `json:"parentsField,omitempty" yaml:"parentsField,omitempty" toml:"parentsField,omitempty"` // code of the field containing the optional parent keys for the entity.
}

// RelationMapping defines the mapping for a relation.
type RelationMapping struct {
	Base               string `json:"base" yaml:"base" toml:"base"`                                                                         // the unique base key for processing a relation.
	SubjectTypeField   string `json:"subjectTypeField,omitempty" yaml:"subjectTypeField,omitempty" toml:"subjectTypeField,omitempty"`       // code of the field containing the subject type. Mutually exclusive with SubjectTypeValue.
	SubjectTypeValue   string `json:"subjectTypeValue,omitempty" yaml:"subjectTypeValue,omitempty" toml:"subjectTypeValue,omitempty"`       // fixed value for the subject type. Mutually exclusive with SubjectTypeField.
	SubjectIdField     string `json:"subjectIdField,omitempty" yaml:"subjectIdField,omitempty" toml:"subjectIdField,omitempty"`             // code of the field containing the subject ID. Mutually exclusive with SubjectIdValue.
	SubjectIdValue     string `json:"subjectIdValue,omitempty" yaml:"subjectIdValue,omitempty" toml:"subjectIdValue,omitempty"`             // fixed value for the subject ID. Mutually exclusive with SubjectIdField.
	PredicateTypeField string `json:"predicateTypeField,omitempty" yaml:"predicateTypeField,omitempty" toml:"predicateTypeField,omitempty"` // code of the field containing the predicate type. Mutually exclusive with PredicateTypeValue.
	PredicateTypeValue string `json:"predicateTypeValue,omitempty" yaml:"predicateTypeValue,omitempty" toml:"predicateTypeValue,omitempty"` // fixed value for the predicate type. Mutually exclusive with PredicateTypeField.
	PredicateIdField   string `json:"predicateIdField,omitempty" yaml:"predicateIdField,omitempty" toml:"predicateIdField,omitempty"`       // code of the field containing the predicate ID. Mutually exclusive with PredicateIdValue.
	PredicateIdValue   string `json:"predicateIdValue,omitempty" yaml:"predicateIdValue,omitempty" toml:"predicateIdValue,omitempty"`       // fixed value for the predicate ID. Mutually exclusive with PredicateIdField.
	ObjectTypeField    string `json:"objectTypeField,omitempty" yaml:"objectTypeField,omitempty" toml:"objectTypeField,omitempty"`          // code of the field containing the object type. Mutually exclusive with ObjectTypeValue.
	ObjectTypeValue    string `json:"objectTypeValue,omitempty" yaml:"objectTypeValue,omitempty" toml:"objectTypeValue,omitempty"`          // fixed value for the object type. Mutually exclusive with ObjectTypeField.
	ObjectIdField      string `json:"objectIdField,omitempty" yaml:"objectIdField,omitempty" toml:"objectIdField,omitempty"`                // code of the field containing the object ID. Mutually exclusive with ObjectIdValue.
	ObjectIdValue      string `json:"objectIdValue,omitempty" yaml:"objectIdValue,omitempty" toml:"objectIdValue,omitempty"`                // fixed value for the object ID. Mutually exclusive with ObjectIdField.
}
