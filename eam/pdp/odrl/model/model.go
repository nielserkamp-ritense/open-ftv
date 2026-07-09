// Package model contains a Go representation of ODRL-AP-NL policies together
// with parsers and serializers for the two concrete syntaxes used by the
// profile: RDF/Turtle (text/turtle) and JSON-LD (application/ld+json).
//
// The model is deliberately a faithful, mostly-lossless projection of the RDF
// graph: every ODRL/AP-NL entity that the profile defines (Policy/Set/Offer/
// Agreement, Permission/Prohibition/Duty, Constraint/LogicalConstraint,
// PolicyArtifact/PolicyBundle) has a corresponding struct, and unknown
// predicates on rules and policies are preserved in an Extra bag so that a
// round-trip does not silently drop information.
package model

// LangString is a (possibly language-tagged) textual value.
type LangString struct {
	Value    string `json:"value"`
	Language string `json:"language,omitempty"`
}

// Value is an operand or object value: either an IRI reference (ID) or a
// literal (Value with optional Language/Datatype).
type Value struct {
	ID       string `json:"id,omitempty"`
	Value    string `json:"value,omitempty"`
	Language string `json:"language,omitempty"`
	Datatype string `json:"datatype,omitempty"`
}

// IsRef reports whether the value is an IRI reference.
func (v Value) IsRef() bool { return v.ID != "" }

// Property captures an otherwise-unmodelled predicate and its values, so that a
// parse→serialize round-trip preserves domain extensions (e.g. brp:medium).
type Property struct {
	Predicate string  `json:"predicate"`
	Values    []Value `json:"values"`
}

// Constraint is an odrl:Constraint or odrl:LogicalConstraint. A plain reference
// to a named constraint is represented with only Ref set.
type Constraint struct {
	UID          string        `json:"uid,omitempty"`
	Ref          string        `json:"ref,omitempty"`
	LeftOperand  string        `json:"leftOperand,omitempty"`
	Operator     string        `json:"operator,omitempty"`
	RightOperand []Value       `json:"rightOperand,omitempty"`
	Labels       []LangString  `json:"labels,omitempty"`
	Comments     []LangString  `json:"comments,omitempty"`
	Waivable     *bool         `json:"waivable,omitempty"`
	LogicalOp    string        `json:"logicalOperator,omitempty"` // odrl:and/or/xone/andSequence IRI.
	Operands     []*Constraint `json:"operands,omitempty"`
}

// Action is an odrl:action value: either a plain IRI (Value) or a structured
// action carrying rdf:value plus refinements / an informedParty (for duties).
type Action struct {
	Value         string        `json:"value,omitempty"`
	Refinements   []*Constraint `json:"refinements,omitempty"`
	InformedParty string        `json:"informedParty,omitempty"`
}

// Duty is an odrl:Duty. A plain reference to a named duty sets only Ref.
type Duty struct {
	UID    string       `json:"uid,omitempty"`
	Ref    string       `json:"ref,omitempty"`
	Action *Action      `json:"action,omitempty"`
	Labels []LangString `json:"labels,omitempty"`
}

// Rule is an odrl:Permission or odrl:Prohibition.
type Rule struct {
	UID         string        `json:"uid,omitempty"`
	Action      *Action       `json:"action,omitempty"`
	Targets     []string      `json:"targets,omitempty"`
	Assigner    string        `json:"assigner,omitempty"`
	Assignee    string        `json:"assignee,omitempty"`
	Constraints []*Constraint `json:"constraints,omitempty"`
	Duties      []*Duty       `json:"duties,omitempty"`
	Extra       []Property    `json:"extra,omitempty"`
}

// Policy is an odrl:Policy — concretely a Set, Offer or Agreement.
type Policy struct {
	UID            string       `json:"uid"`
	Type           string       `json:"type"` // odrl class IRI (Set/Offer/Agreement).
	Profile        string       `json:"profile,omitempty"`
	Profiles       []string     `json:"profiles,omitempty"` // full odrl:profile set (e.g. apnl + geonl).
	Titles         []LangString `json:"titles,omitempty"`
	Descriptions   []LangString `json:"descriptions,omitempty"`
	Publisher      string       `json:"publisher,omitempty"`
	Issued         string       `json:"issued,omitempty"`
	Assigner       string       `json:"assigner,omitempty"`
	Assignee       string       `json:"assignee,omitempty"`
	Instantiates   string       `json:"instantiates,omitempty"`
	WasDerivedFrom string       `json:"wasDerivedFrom,omitempty"`
	WasRevisionOf  string       `json:"wasRevisionOf,omitempty"`
	Permissions    []*Rule      `json:"permissions,omitempty"`
	Prohibitions   []*Rule      `json:"prohibitions,omitempty"`
	Obligations    []*Duty      `json:"obligations,omitempty"`
	Extra          []Property   `json:"extra,omitempty"`
}

// Artifact is an apnl:PolicyArtifact (RegoModule / CedarPolicySet / OpenFGAModel).
type Artifact struct {
	UID            string       `json:"uid"`
	Type           string       `json:"type"` // apnl class IRI.
	Titles         []LangString `json:"titles,omitempty"`
	Descriptions   []LangString `json:"descriptions,omitempty"`
	Format         string       `json:"format,omitempty"`
	DownloadURL    string       `json:"downloadURL,omitempty"`
	Entrypoint     string       `json:"entrypoint,omitempty"`
	SHA256         string       `json:"sha256,omitempty"`
	WasRevisionOf  string       `json:"wasRevisionOf,omitempty"`
	WasDerivedFrom string       `json:"wasDerivedFrom,omitempty"`
}

// Bundle is an apnl:PolicyBundle.
type Bundle struct {
	UID           string       `json:"uid"`
	Titles        []LangString `json:"titles,omitempty"`
	Descriptions  []LangString `json:"descriptions,omitempty"`
	Publisher     string       `json:"publisher,omitempty"`
	Issued        string       `json:"issued,omitempty"`
	Format        string       `json:"format,omitempty"`
	DownloadURL   string       `json:"downloadURL,omitempty"`
	Entrypoint    string       `json:"entrypoint,omitempty"`
	SHA256        string       `json:"sha256,omitempty"`
	Bundles       []string     `json:"bundles,omitempty"` // artifact IRIs.
	WasRevisionOf string       `json:"wasRevisionOf,omitempty"`
}

// Document is a parsed ODRL-AP-NL graph: the ODRL/AP-NL entities of interest.
type Document struct {
	Policies    []*Policy     `json:"policies,omitempty"`
	Artifacts   []*Artifact   `json:"artifacts,omitempty"`
	Bundles     []*Bundle     `json:"bundles,omitempty"`
	Constraints []*Constraint `json:"constraints,omitempty"` // named (top-level) constraints.
	Duties      []*Duty       `json:"duties,omitempty"`      // named (top-level) duties.
}

// Policy returns the policy with the given uid, or nil.
func (d *Document) Policy(uid string) *Policy {
	for _, p := range d.Policies {
		if p.UID == uid {
			return p
		}
	}
	return nil
}

// Artifact returns the artifact with the given uid, or nil.
func (d *Document) Artifact(uid string) *Artifact {
	for _, a := range d.Artifacts {
		if a.UID == uid {
			return a
		}
	}
	return nil
}

func boolPtr(b bool) *bool { return &b }
