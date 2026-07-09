package odrl_geo

// The geo policy model. It is a compact, evaluation-oriented projection of an
// ODRL-Geo-NL policy that preserves exactly the geo-specific information the
// shared ODRL-AP-NL parser drops (geonl:property, odrl:rightOperandReference,
// geonl:Area / geonl:LayerSelection). It is built by the graph parser
// (parse.go) from the raw policy source kept in the PAP envelope.

// geoPolicy is one ODRL Offer or Agreement enriched for geo evaluation.
type geoPolicy struct {
	UID          string
	Type         string // odrl Offer / Agreement IRI.
	Profile      []string
	Assigner     string
	Assignee     string
	Instantiates string
	DefaultCRS   string
	Targets      []string // union of rule targets (for logging/diagnostics).

	Permissions  []*geoRule
	Prohibitions []*geoRule
}

// isAgreement reports whether the policy is an odrl:Agreement.
func (p *geoPolicy) isAgreement() bool { return p.Type == odrlNS+"Agreement" }

// geoRule is a permission or prohibition.
type geoRule struct {
	Action      string // resolved odrl action IRI (e.g. odrl:read).
	Targets     []string
	Purpose     *purposeRefinement // purpose refinement on the action, if any.
	Constraints []*geoConstraint
	Duties      []*geoDuty
}

// purposeRefinement is an odrl:purpose refinement on a rule action (trap C).
type purposeRefinement struct {
	Operator string // odrl:isA / isAnyOf / eq.
	Values   []string
}

// geoConstraint is a single feature-property or spatial constraint.
type geoConstraint struct {
	LeftOperand string // geonl:featureProperty / geonl:featureGeometry / attribute URI.
	Property    string // geonl:property value (attribute URI or column name).
	Operator    string // odrl operator or geof: spatial operator.

	// feature-property right operands (literals). For isAnyOf/isAllOf there
	// may be several.
	RightLiterals []literal

	// spatial right operands: an inline geometry literal, or a reference to a
	// named Area / LayerSelection resolved via the document.
	RightGeometry string // inline geo:wktLiteral (may carry a CRS prefix).
	RightRef      string // odrl:rightOperandReference IRI.

	Waivable bool
}

// isSpatial reports whether the constraint is a spatial (featureGeometry) one.
func (c *geoConstraint) isSpatial() bool { return c.LeftOperand == geonlFeatureGeometry }

// literal is a typed RDF literal value.
type literal struct {
	Value    string
	Datatype string
}

// geoDuty is a "ja, mits" duty mapped to an AuthZEN obligation.
type geoDuty struct {
	Action      string       // geonl: measure IRI.
	Refinements []refinement // flattened to obligation properties.
}

// refinement is a duty-action refinement (geonl:schaalnoemer >= 50000, ...).
type refinement struct {
	Key   string // localname of the refinement left operand.
	Value string
}

// area is a named geonl:Area with its geometry.
type area struct {
	UID string
	WKT string // geo:asWKT literal (may carry a CRS prefix).
}

// layerSelection is a geonl:LayerSelection: geometries from another layer,
// optionally filtered by an attribute constraint.
type layerSelection struct {
	UID       string
	FromLayer string
	Where     *geoConstraint // attribute filter (featureProperty pattern).
}

// document bundles the policies of one PAP entry plus the shared geo entities
// they may reference.
type document struct {
	Policies   []*geoPolicy
	Areas      map[string]*area
	Selections map[string]*layerSelection
	// accessURL maps a target IRI (dcat:Distribution) to its dcat:accessURL, so
	// a rule target can be matched against resource.properties.layer.
	AccessURL map[string]string
}
