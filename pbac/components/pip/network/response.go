package network

// Response contains the parameters used to decode a response when pulling external data.
//
// Supported encodings are JSON, YAML, TOML, RDF/Turtle and RDF/JSON-LD.
// The server should send the Content-Type header to indicate the format of the response body.
// If it doesn't, the content type will be determined by looking at the first bytes of the body.
// This may fail for a variety of reasons, so it is recommended to always send the appropriate Content-Type header.
//
// An RDF based response body **must** adhere to the FTV ontology.
//
// The key for each map indicates the base key where objects of that type can be found in the response.
// E.g. the response is decoded into a recursive tree of key/value pairs.
// A base key contains the concatenated key values to the base of an object or array of objects of that type.
// Key values are concatenated with a dot, as common in object-oriented languages; e.g. "data.network.attributes".
// Keys that contain one or more dots, must be enclosed in double or single quotes; e.g. "data.'net.addresses'.attributes".
type Response struct {
	Attributes map[string]*AttributeObject `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"` // the base(s) for processing attributes.
	Entities   map[string]*EntityObject    `json:"entities,omitempty" yaml:"entities,omitempty" toml:"entities,omitempty"`       // the base(s) for processing entities.
	Relations  map[string]*RelationObject  `json:"relations,omitempty" yaml:"relations,omitempty" toml:"relations,omitempty"`    // the base(s) for processing relations.
}

// AttributeObject defines how to transform an attribute from the response.
type AttributeObject struct {
	Base      string `json:"base" yaml:"base" toml:"base"`                                              // the unique base key for processing this type of attribute.
	KeyCode   string `json:"keyCode,omitempty" yaml:"keyCode,omitempty" toml:"keyCode,omitempty"`       // code of the field containing the attribute key. Mutually exclusive with KeyValue.
	KeyValue  string `json:"keyValue,omitempty" yaml:"keyValue,omitempty" toml:"keyValue,omitempty"`    // value for the attribute key. Mutually exclusive with KeyCode.
	ValueCode string `json:"valueCode,omitempty" yaml:"valueCode,omitempty" toml:"valueCode,omitempty"` // code of the field containing the attribute value. Mutually exclusive with ValueAsIs.
	ValueAsIs bool   `json:"valueAsIs,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"` // indicates the response object itself is the value for the attribute. Mutually exclusive with ValueCode.
	TypeCode  string `json:"typeCode,omitempty" yaml:"typeCode,omitempty" toml:"typeCode,omitempty"`    // code of the field containing the type of the attribute. Mutually exclusive with TypeValue.
	TypeValue string `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"` // type of the attribute. Mutually exclusive with TypeCode.
}

// EntityObject defines how to transform an entity from the response.
type EntityObject struct {
	Base        string                      `json:"base" yaml:"base" toml:"base"`                                                    // the unique base key for processing this type of entity.
	TypeCode    string                      `json:"typeCode,omitempty" yaml:"typeCode,omitempty" toml:"typeCode,omitempty"`          // code of the field containing the entity type. Mutually exclusive with TypeValue.
	TypeValue   string                      `json:"typeValue,omitempty" yaml:"typeValue,omitempty" toml:"typeValue,omitempty"`       // value for the entity type. Mutually exclusive with TypeCode.
	IdCode      string                      `json:"idCode,omitempty" yaml:"idCode,omitempty" toml:"idCode,omitempty"`                // code of the field containing the entity ID. Mutually exclusive with IdValue and IdFromValue.
	IdValue     string                      `json:"idValue,omitempty" yaml:"idValue,omitempty" toml:"idValue,omitempty"`             // value for the entity ID. Mutually exclusive with IdCode and IdFromValue.
	IdFromValue bool                        `json:"idFromValue,omitempty" yaml:"valueAsIs,omitempty" toml:"valueAsIs,omitempty"`     // indicates the response object itself is the value for the ID. Mutually exclusive with IdValue and IdCode.
	Attributes  map[string]*AttributeObject `json:"attributes,omitempty" yaml:"attributes,omitempty" toml:"attributes,omitempty"`    // relative base(s) for processing attributes for the entity.
	ParentsCode string                      `json:"parentsCode,omitempty" yaml:"parentsCode,omitempty" toml:"parentsCode,omitempty"` // code of the field containing the optional parent keys for the entity.
}

// RelationObject defines how to transform a relation from the response.
type RelationObject struct {
	Base               string `json:"base" yaml:"base" toml:"base"`                                                                         // the unique base key for processing this type of relation.
	SubjectTypeCode    string `json:"subjectTypeCode,omitempty" yaml:"subjectTypeCode,omitempty" toml:"subjectTypeCode,omitempty"`          // code of the field containing the subject type. Mutually exclusive with SubjectTypeValue.
	SubjectTypeValue   string `json:"subjectTypeValue,omitempty" yaml:"subjectTypeValue,omitempty" toml:"subjectTypeValue,omitempty"`       // value for the subject type. Mutually exclusive with SubjectTypeCode.
	SubjectIdCode      string `json:"subjectIdCode,omitempty" yaml:"subjectIdCode,omitempty" toml:"subjectIdCode,omitempty"`                // code of the field containing the subject ID. Mutually exclusive with SubjectIdValue.
	SubjectIdValue     string `json:"subjectIdValue,omitempty" yaml:"subjectIdValue,omitempty" toml:"subjectIdValue,omitempty"`             // value for the subject ID. Mutually exclusive with SubjectIdCode.
	PredicateTypeCode  string `json:"predicateTypeCode,omitempty" yaml:"predicateTypeCode,omitempty" toml:"predicateTypeCode,omitempty"`    // code of the field containing the predicate type. Mutually exclusive with PredicateTypeValue.
	PredicateTypeValue string `json:"predicateTypeValue,omitempty" yaml:"predicateTypeValue,omitempty" toml:"predicateTypeValue,omitempty"` // value for the predicate type. Mutually exclusive with PredicateTypeCode.
	PredicateIdCode    string `json:"predicateIdCode,omitempty" yaml:"predicateIdCode,omitempty" toml:"predicateIdCode,omitempty"`          // code of the field containing the predicate ID. Mutually exclusive with PredicateIdValue.
	PredicateIdValue   string `json:"predicateIdValue,omitempty" yaml:"predicateIdValue,omitempty" toml:"predicateIdValue,omitempty"`       // value for the predicate ID. Mutually exclusive with PredicateIdCode.
	ObjectTypeCode     string `json:"objectTypeCode,omitempty" yaml:"objectTypeCode,omitempty" toml:"objectTypeCode,omitempty"`             // code of the field containing the object type. Mutually exclusive with ObjectTypeValue.
	ObjectTypeValue    string `json:"objectTypeValue,omitempty" yaml:"objectTypeValue,omitempty" toml:"objectTypeValue,omitempty"`          // value for the object type. Mutually exclusive with ObjectTypeCode.
	ObjectIdCode       string `json:"objectIdCode,omitempty" yaml:"objectIdCode,omitempty" toml:"objectIdCode,omitempty"`                   // code of the field containing the object ID. Mutually exclusive with ObjectIdValue.
	ObjectIdValue      string `json:"objectIdValue,omitempty" yaml:"objectIdValue,omitempty" toml:"objectIdValue,omitempty"`                // value for the object ID. Mutually exclusive with ObjectIdCode.
}
