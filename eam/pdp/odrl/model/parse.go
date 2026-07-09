package model

import (
	"io"
	"strings"

	"github.com/deiu/rdf2go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/rdf"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/turtle"
)

// Parse reads an ODRL-AP-NL document from r. The mime type must be one of
// "text/turtle" or "application/ld+json".
func Parse(r io.Reader, mime string) (*Document, error) {
	g, err := turtle.Load(r, mime)
	if err != nil {
		return nil, err
	}
	return FromGraph(g), nil
}

// FromGraph projects an already-parsed RDF graph into a Document.
func FromGraph(g *rdf2go.Graph) *Document {
	p := &parser{g: g}
	return p.run()
}

type parser struct {
	g *rdf2go.Graph
}

func res(uri string) rdf2go.Term { return rdf2go.NewResource(uri) }

func (p *parser) run() *Document {
	d := &Document{}

	for _, t := range []string{rdf.ODRLSet, rdf.ODRLOffer, rdf.ODRLAgreement} {
		for _, tr := range p.g.All(nil, res(rdf.Type), res(t)) {
			d.Policies = append(d.Policies, p.policy(tr.Subject, t))
		}
	}

	for _, t := range []string{APNLPolicyArtifact, APNLRegoModule, APNLCedarPolicySet, APNLOpenFGAModel} {
		for _, tr := range p.g.All(nil, res(rdf.Type), res(t)) {
			d.Artifacts = append(d.Artifacts, p.artifact(tr.Subject, t))
		}
	}

	for _, tr := range p.g.All(nil, res(rdf.Type), res(APNLPolicyBundle)) {
		d.Bundles = append(d.Bundles, p.bundle(tr.Subject))
	}

	// Named (IRI) constraints and duties that are declared at top level.
	for _, tr := range p.g.All(nil, res(rdf.Type), res(rdf.ODRLConstraint)) {
		if _, ok := tr.Subject.(*rdf2go.Resource); ok {
			d.Constraints = append(d.Constraints, p.constraint(tr.Subject))
		}
	}
	for _, tr := range p.g.All(nil, res(rdf.Type), res(rdf.ODRLDuty)) {
		if _, ok := tr.Subject.(*rdf2go.Resource); ok {
			d.Duties = append(d.Duties, p.duty(tr.Subject))
		}
	}

	return d
}

func (p *parser) policy(sub rdf2go.Term, typ string) *Policy {
	pol := &Policy{UID: iri(sub), Type: typ}
	if u := p.oneRes(sub, rdf.ODRLIdentifier); u != "" {
		pol.UID = u
	}
	pol.Profiles = p.allRes(sub, rdf.ODRLProfile)
	pol.Profile = primaryProfile(pol.Profiles)
	pol.Titles = p.langs(sub, DCTTitle)
	pol.Descriptions = p.langs(sub, DCTDescription)
	pol.Publisher = p.oneRes(sub, DCTPublisher)
	pol.Issued = p.oneLit(sub, DCTIssued)
	pol.Assigner = p.oneRes(sub, rdf.ODRLAssigner)
	pol.Assignee = p.oneRes(sub, rdf.ODRLAssignee)
	pol.Instantiates = p.oneRes(sub, APNLInstantiates)
	pol.WasDerivedFrom = p.oneRes(sub, PROVWasDerivedFrom)
	pol.WasRevisionOf = p.oneRes(sub, PROVWasRevisionOf)

	for _, tr := range p.g.All(sub, res(rdf.ODRLHasPermission), nil) {
		pol.Permissions = append(pol.Permissions, p.rule(tr.Object))
	}
	for _, tr := range p.g.All(sub, res(rdf.ODRLHasProhibition), nil) {
		pol.Prohibitions = append(pol.Prohibitions, p.rule(tr.Object))
	}
	for _, tr := range p.g.All(sub, res(rdf.ODRLObligation), nil) {
		pol.Obligations = append(pol.Obligations, p.dutyRef(tr.Object))
	}

	pol.Extra = p.extra(sub, policyKnown)
	return pol
}

func (p *parser) rule(sub rdf2go.Term) *Rule {
	r := &Rule{UID: iriIfResource(sub)}
	if a := p.action(sub); a != nil {
		r.Action = a
	}
	for _, tr := range p.g.All(sub, res(ODRLTargetProp), nil) {
		r.Targets = append(r.Targets, iri(tr.Object))
	}
	r.Assigner = p.oneRes(sub, rdf.ODRLAssigner)
	r.Assignee = p.oneRes(sub, rdf.ODRLAssignee)
	for _, tr := range p.g.All(sub, res(rdf.ODRLHasConstraint), nil) {
		r.Constraints = append(r.Constraints, p.constraintRef(tr.Object))
	}
	for _, tr := range p.g.All(sub, res(rdf.ODRLHasDuty), nil) {
		r.Duties = append(r.Duties, p.dutyRef(tr.Object))
	}
	r.Extra = p.extra(sub, ruleKnown)
	return r
}

// action parses the odrl:action of a rule/duty. It returns nil when absent.
func (p *parser) action(sub rdf2go.Term) *Action {
	tr := p.g.One(sub, res(rdf.ODRLHasAction), nil)
	if tr == nil {
		return nil
	}
	obj := tr.Object
	if r, ok := obj.(*rdf2go.Resource); ok {
		return &Action{Value: r.URI}
	}
	// structured action (blank node): rdf:value + refinement / informedParty.
	a := &Action{Value: p.oneRes(obj, rdf.Value)}
	for _, rt := range p.g.All(obj, res(rdf.ODRLRefinement), nil) {
		a.Refinements = append(a.Refinements, p.constraintRef(rt.Object))
	}
	a.InformedParty = p.oneRes(obj, ODRLInformedParty)
	return a
}

// constraintRef resolves a constraint object on a rule/refinement. IRI-named
// constraints become references: their full definition is collected once at
// document level (Document.Constraints), unless the definition is untyped —
// in which case it would be lost and is therefore inlined.
func (p *parser) constraintRef(obj rdf2go.Term) *Constraint {
	if r, ok := obj.(*rdf2go.Resource); ok {
		if p.g.One(r, res(rdf.Type), res(rdf.ODRLConstraint)) != nil ||
			p.g.One(r, res(rdf.Type), res(rdf.ODRLLogicalConstraint)) != nil {
			return &Constraint{Ref: r.URI}
		}
		if p.g.One(r, res(rdf.ODRLHasLeftOperand), nil) == nil &&
			p.g.One(r, res(rdf.ODRLOperand), nil) == nil {
			return &Constraint{Ref: r.URI}
		}
	}
	return p.constraint(obj)
}

func (p *parser) constraint(sub rdf2go.Term) *Constraint {
	c := &Constraint{UID: iriIfResource(sub)}
	c.LeftOperand = p.oneRes(sub, rdf.ODRLHasLeftOperand)
	c.Operator = p.oneRes(sub, rdf.ODRLOperator)
	for _, tr := range p.g.All(sub, res(rdf.ODRLHasRightOperand), nil) {
		c.RightOperand = append(c.RightOperand, termValue(tr.Object))
	}
	c.Labels = p.langs(sub, RDFSLabel)
	c.Comments = p.langs(sub, RDFSComment)
	if b, ok := p.oneBool(sub, APNLWaivable); ok {
		c.Waivable = boolPtr(b)
	}
	// logical constraint operators.
	for _, op := range []string{rdf.ODRLAnd, rdf.ODRLAndSequence, rdf.ODRLOr, rdf.ODRLOnlyOne} {
		if tr := p.g.One(sub, res(op), nil); tr != nil {
			c.LogicalOp = op
			for _, m := range p.list(tr.Object) {
				c.Operands = append(c.Operands, p.constraintRef(m))
			}
		}
	}
	return c
}

func (p *parser) duty(sub rdf2go.Term) *Duty {
	d := &Duty{UID: iriIfResource(sub)}
	d.Action = p.action(sub)
	d.Labels = p.langs(sub, RDFSLabel)
	return d
}

// dutyRef resolves a duty object: IRI-named duties become references (their
// definition is collected in Document.Duties when typed odrl:Duty), untyped
// named duties are inlined so their triples survive.
func (p *parser) dutyRef(obj rdf2go.Term) *Duty {
	if r, ok := obj.(*rdf2go.Resource); ok {
		if p.g.One(r, res(rdf.Type), res(rdf.ODRLDuty)) != nil {
			return &Duty{Ref: r.URI}
		}
		if p.g.One(r, res(rdf.ODRLHasAction), nil) == nil {
			return &Duty{Ref: r.URI}
		}
	}
	return p.duty(obj)
}

func (p *parser) artifact(sub rdf2go.Term, typ string) *Artifact {
	a := &Artifact{UID: iri(sub), Type: typ}
	a.Titles = p.langs(sub, DCTTitle)
	a.Descriptions = p.langs(sub, DCTDescription)
	a.Format = p.oneLit(sub, DCTFormat)
	a.DownloadURL = p.oneRes(sub, DCATDownloadURL)
	a.Entrypoint = p.oneLit(sub, APNLEntrypoint)
	a.SHA256 = p.oneLit(sub, APNLSha256)
	a.WasRevisionOf = p.oneRes(sub, PROVWasRevisionOf)
	a.WasDerivedFrom = p.oneRes(sub, PROVWasDerivedFrom)
	return a
}

func (p *parser) bundle(sub rdf2go.Term) *Bundle {
	b := &Bundle{UID: iri(sub)}
	b.Titles = p.langs(sub, DCTTitle)
	b.Descriptions = p.langs(sub, DCTDescription)
	b.Publisher = p.oneRes(sub, DCTPublisher)
	b.Issued = p.oneLit(sub, DCTIssued)
	b.Format = p.oneLit(sub, DCTFormat)
	b.DownloadURL = p.oneRes(sub, DCATDownloadURL)
	b.Entrypoint = p.oneLit(sub, APNLEntrypoint)
	b.SHA256 = p.oneLit(sub, APNLSha256)
	b.WasRevisionOf = p.oneRes(sub, PROVWasRevisionOf)
	for _, tr := range p.g.All(sub, res(APNLBundles), nil) {
		b.Bundles = append(b.Bundles, iri(tr.Object))
	}
	return b
}

// extra collects predicates not in the known set into Property bags.
func (p *parser) extra(sub rdf2go.Term, known map[string]bool) []Property {
	byPred := map[string][]Value{}
	var order []string
	for tr := range p.g.IterTriples() {
		if !tr.Subject.Equal(sub) {
			continue
		}
		pred := iri(tr.Predicate)
		if known[pred] {
			continue
		}
		if _, ok := byPred[pred]; !ok {
			order = append(order, pred)
		}
		byPred[pred] = append(byPred[pred], termValue(tr.Object))
	}
	var out []Property
	for _, pred := range order {
		out = append(out, Property{Predicate: pred, Values: byPred[pred]})
	}
	return out
}

// list walks an RDF collection (rdf:first/rdf:rest) starting at head.
func (p *parser) list(head rdf2go.Term) []rdf2go.Term {
	var out []rdf2go.Term
	nilRes := rdf.URIRDF + "nil"
	for head != nil {
		if r, ok := head.(*rdf2go.Resource); ok && r.URI == nilRes {
			break
		}
		first := p.g.One(head, res(rdf.URIRDF+"first"), nil)
		if first == nil {
			// not a list: treat head itself as the single member.
			out = append(out, head)
			break
		}
		out = append(out, first.Object)
		rest := p.g.One(head, res(rdf.URIRDF+"rest"), nil)
		if rest == nil {
			break
		}
		head = rest.Object
	}
	return out
}

// --- small term helpers ---

func (p *parser) oneRes(sub rdf2go.Term, pred string) string {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return ""
	}
	return iri(tr.Object)
}

func (p *parser) allRes(sub rdf2go.Term, pred string) []string {
	var out []string
	for _, tr := range p.g.All(sub, res(pred), nil) {
		if v := iri(tr.Object); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// primaryProfile picks the single representative profile for Policy.Profile:
// the ODRL-AP-NL base profile when present, otherwise the first declared one.
func primaryProfile(profiles []string) string {
	for _, prof := range profiles {
		if prof == APNLProfile {
			return prof
		}
	}
	if len(profiles) > 0 {
		return profiles[0]
	}
	return ""
}

func (p *parser) oneLit(sub rdf2go.Term, pred string) string {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return ""
	}
	return tr.Object.RawValue()
}

func (p *parser) oneBool(sub rdf2go.Term, pred string) (bool, bool) {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return false, false
	}
	return strings.EqualFold(tr.Object.RawValue(), "true"), true
}

func (p *parser) langs(sub rdf2go.Term, pred string) []LangString {
	var out []LangString
	for _, tr := range p.g.All(sub, res(pred), nil) {
		if l, ok := tr.Object.(*rdf2go.Literal); ok {
			out = append(out, LangString{Value: l.Value, Language: l.Language})
		} else {
			out = append(out, LangString{Value: tr.Object.RawValue()})
		}
	}
	return out
}

func termValue(t rdf2go.Term) Value {
	switch x := t.(type) {
	case *rdf2go.Resource:
		return Value{ID: x.URI}
	case *rdf2go.Literal:
		v := Value{Value: x.Value, Language: x.Language}
		if x.Datatype != nil {
			v.Datatype = x.Datatype.RawValue()
		}
		return v
	case *rdf2go.BlankNode:
		return Value{ID: "_:" + x.ID}
	default:
		return Value{Value: t.RawValue()}
	}
}

func iri(t rdf2go.Term) string {
	switch x := t.(type) {
	case *rdf2go.Resource:
		return x.URI
	case *rdf2go.BlankNode:
		return "_:" + x.ID
	default:
		return t.RawValue()
	}
}

func iriIfResource(t rdf2go.Term) string {
	if r, ok := t.(*rdf2go.Resource); ok {
		return r.URI
	}
	return ""
}

// known-predicate sets used to compute the Extra bag.
var policyKnown = toSet(
	rdf.Type, rdf.ODRLIdentifier, rdf.ODRLProfile, DCTTitle, DCTDescription,
	DCTPublisher, DCTIssued, rdf.ODRLAssigner, rdf.ODRLAssignee, APNLInstantiates,
	PROVWasDerivedFrom, PROVWasRevisionOf, rdf.ODRLHasPermission,
	rdf.ODRLHasProhibition, rdf.ODRLObligation,
)

var ruleKnown = toSet(
	rdf.Type, rdf.ODRLHasAction, rdf.ODRLTarget, rdf.ODRLAssigner,
	rdf.ODRLAssignee, rdf.ODRLHasConstraint, rdf.ODRLHasDuty,
)

func toSet(keys ...string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
