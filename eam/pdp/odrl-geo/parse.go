package odrl_geo

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/deiu/rdf2go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/turtle"
)

const dcatAccessURL = "http://www.w3.org/ns/dcat#accessURL"

// loadDocument decodes a PAP "odrl" policy envelope (as imported by
// eam/pdp/odrl/importer) and builds the geo evaluation model from the raw
// source graph. It reuses the shared parser's Envelope to obtain the original
// bytes, then re-parses the graph with geo awareness.
func loadDocument(envelopeJSON []byte) (*document, error) {
	env, err := model.ParseEnvelope(envelopeJSON)
	if err != nil {
		return nil, fmt.Errorf("odrl-geo: %w", err)
	}
	raw, err := env.RawSource()
	if err != nil {
		return nil, fmt.Errorf("odrl-geo: cannot decode envelope source: %w", err)
	}
	g, err := turtle.Load(bytes.NewReader(raw), env.MimeType)
	if err != nil {
		return nil, fmt.Errorf("odrl-geo: cannot parse policy graph: %w", err)
	}
	return fromGraph(g), nil
}

// fromGraph projects an RDF graph into the geo document model.
func fromGraph(g *rdf2go.Graph) *document {
	p := &gparser{g: g}
	return p.run()
}

type gparser struct {
	g *rdf2go.Graph
}

func (p *gparser) run() *document {
	d := &document{
		Areas:      map[string]*area{},
		Selections: map[string]*layerSelection{},
		AccessURL:  map[string]string{},
	}

	// dcat:accessURL of distributions, so targets can match resource layers.
	for _, tr := range p.g.All(nil, res(dcatAccessURL), nil) {
		d.AccessURL[iri(tr.Subject)] = iri(tr.Object)
	}

	// named Areas.
	for _, tr := range p.g.All(nil, res(rdfNS+"type"), res(geonlArea)) {
		a := &area{UID: iri(tr.Subject)}
		if geom := p.g.One(tr.Subject, res(geoHasGeometry), nil); geom != nil {
			a.WKT = p.oneLit(geom.Object, geoAsWKT)
		}
		d.Areas[a.UID] = a
	}

	// named LayerSelections.
	for _, tr := range p.g.All(nil, res(rdfNS+"type"), res(geonlLayerSelection)) {
		ls := &layerSelection{UID: iri(tr.Subject)}
		ls.FromLayer = p.oneRes(tr.Subject, geonlFromLayer)
		if w := p.g.One(tr.Subject, res(geonlWhere), nil); w != nil {
			ls.Where = p.constraint(w.Object)
		}
		d.Selections[ls.UID] = ls
	}

	// policies: Offers and Agreements (Sets treated the same).
	for _, typ := range []string{odrlNS + "Offer", odrlNS + "Agreement", odrlNS + "Set"} {
		for _, tr := range p.g.All(nil, res(rdfNS+"type"), res(typ)) {
			d.Policies = append(d.Policies, p.policy(tr.Subject, typ))
		}
	}

	return d
}

func (p *gparser) policy(sub rdf2go.Term, typ string) *geoPolicy {
	pol := &geoPolicy{UID: iri(sub), Type: typ}
	if u := p.oneRes(sub, odrlNS+"uid"); u != "" {
		pol.UID = u
	}
	for _, tr := range p.g.All(sub, res(odrlNS+"profile"), nil) {
		pol.Profile = append(pol.Profile, iri(tr.Object))
	}
	pol.Assigner = p.oneRes(sub, odrlNS+"assigner")
	pol.Assignee = p.oneRes(sub, odrlNS+"assignee")
	pol.Instantiates = p.oneRes(sub, "https://standaarden.overheid.nl/odrl-ap-nl/instantiates")
	pol.DefaultCRS = p.oneRes(sub, geonlDefaultCRS)

	for _, tr := range p.g.All(sub, res(odrlNS+"permission"), nil) {
		r := p.rule(tr.Object)
		pol.Permissions = append(pol.Permissions, r)
		pol.Targets = append(pol.Targets, r.Targets...)
	}
	for _, tr := range p.g.All(sub, res(odrlNS+"prohibition"), nil) {
		r := p.rule(tr.Object)
		pol.Prohibitions = append(pol.Prohibitions, r)
		pol.Targets = append(pol.Targets, r.Targets...)
	}
	return pol
}

func (p *gparser) rule(sub rdf2go.Term) *geoRule {
	r := &geoRule{}
	r.Action, r.Purpose = p.action(sub)

	for _, tr := range p.g.All(sub, res(odrlNS+"target"), nil) {
		r.Targets = append(r.Targets, iri(tr.Object))
	}
	for _, tr := range p.g.All(sub, res(odrlConstraintProp), nil) {
		if c := p.constraint(tr.Object); c != nil {
			r.Constraints = append(r.Constraints, c)
		}
	}
	for _, tr := range p.g.All(sub, res(odrlNS+"duty"), nil) {
		if dty := p.duty(tr.Object); dty != nil {
			r.Duties = append(r.Duties, dty)
		}
	}
	return r
}

// action resolves a rule's odrl:action into its IRI plus an optional purpose
// refinement (leftOperand odrl:purpose).
func (p *gparser) action(sub rdf2go.Term) (string, *purposeRefinement) {
	tr := p.g.One(sub, res(odrlActionProp), nil)
	if tr == nil {
		return "", nil
	}
	if r, ok := tr.Object.(*rdf2go.Resource); ok {
		return r.URI, nil
	}
	// structured action (blank node): rdf:value + refinements.
	actionIRI := p.oneRes(tr.Object, rdfValue)
	var purpose *purposeRefinement
	for _, rt := range p.g.All(tr.Object, res(odrlRefinement), nil) {
		c := p.constraint(rt.Object)
		if c != nil && c.LeftOperand == odrlPurpose {
			pr := &purposeRefinement{Operator: c.Operator}
			for _, l := range c.RightLiterals {
				if l.Value != "" {
					pr.Values = append(pr.Values, l.Value)
				}
			}
			if c.RightRef != "" {
				pr.Values = append(pr.Values, c.RightRef)
			}
			purpose = pr
		}
	}
	return actionIRI, purpose
}

func (p *gparser) constraint(sub rdf2go.Term) *geoConstraint {
	c := &geoConstraint{}
	c.LeftOperand = p.oneRes(sub, odrlLeftOperand)
	c.Operator = p.oneRes(sub, odrlOperator)
	c.RightRef = p.oneRes(sub, odrlRightOperandReference)

	// geonl:property may be a resource (attribute URI) or a literal (column).
	if pt := p.g.One(sub, res(geonlProperty), nil); pt != nil {
		c.Property = termText(pt.Object)
	}

	for _, tr := range p.g.All(sub, res(odrlRightOperand), nil) {
		for _, v := range p.expandList(tr.Object) {
			lit := termLiteral(v)
			if strings.HasSuffix(lit.Datatype, "wktLiteral") || strings.HasSuffix(lit.Datatype, "geoJSONLiteral") {
				c.RightGeometry = lit.Value
			} else {
				c.RightLiterals = append(c.RightLiterals, lit)
			}
		}
	}

	if b, ok := p.oneBool(sub, "https://standaarden.overheid.nl/odrl-ap-nl/waivable"); ok {
		c.Waivable = b
	}
	if c.LeftOperand == "" && c.Operator == "" && c.RightRef == "" && len(c.RightLiterals) == 0 {
		return nil
	}
	return c
}

func (p *gparser) duty(sub rdf2go.Term) *geoDuty {
	tr := p.g.One(sub, res(odrlActionProp), nil)
	if tr == nil {
		return nil
	}
	d := &geoDuty{}
	if r, ok := tr.Object.(*rdf2go.Resource); ok {
		d.Action = r.URI
		return d
	}
	d.Action = p.oneRes(tr.Object, rdfValue)
	for _, rt := range p.g.All(tr.Object, res(odrlRefinement), nil) {
		c := p.constraint(rt.Object)
		if c == nil {
			continue
		}
		val := ""
		if len(c.RightLiterals) > 0 {
			val = c.RightLiterals[0].Value
		} else if c.RightRef != "" {
			val = c.RightRef
		}
		d.Refinements = append(d.Refinements, refinement{Key: localName(c.LeftOperand), Value: val})
	}
	return d
}

// expandList returns the members of an RDF collection, or the term itself when
// it is not a list.
func (p *gparser) expandList(head rdf2go.Term) []rdf2go.Term {
	first := p.g.One(head, res(rdfNS+"first"), nil)
	if first == nil {
		return []rdf2go.Term{head}
	}
	var out []rdf2go.Term
	nilRes := rdfNS + "nil"
	for head != nil {
		if r, ok := head.(*rdf2go.Resource); ok && r.URI == nilRes {
			break
		}
		f := p.g.One(head, res(rdfNS+"first"), nil)
		if f == nil {
			break
		}
		out = append(out, f.Object)
		rest := p.g.One(head, res(rdfNS+"rest"), nil)
		if rest == nil {
			break
		}
		head = rest.Object
	}
	return out
}

// --- term helpers ---

func res(uri string) rdf2go.Term { return rdf2go.NewResource(uri) }

func (p *gparser) oneRes(sub rdf2go.Term, pred string) string {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return ""
	}
	return iri(tr.Object)
}

func (p *gparser) oneLit(sub rdf2go.Term, pred string) string {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return ""
	}
	return termText(tr.Object)
}

func (p *gparser) oneBool(sub rdf2go.Term, pred string) (bool, bool) {
	tr := p.g.One(sub, res(pred), nil)
	if tr == nil {
		return false, false
	}
	return strings.EqualFold(termText(tr.Object), "true"), true
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

func termText(t rdf2go.Term) string {
	switch x := t.(type) {
	case *rdf2go.Resource:
		return x.URI
	case *rdf2go.Literal:
		return x.Value
	default:
		return t.RawValue()
	}
}

func termLiteral(t rdf2go.Term) literal {
	switch x := t.(type) {
	case *rdf2go.Literal:
		l := literal{Value: x.Value}
		if x.Datatype != nil {
			l.Datatype = x.Datatype.RawValue()
		}
		return l
	case *rdf2go.Resource:
		return literal{Value: x.URI}
	default:
		return literal{Value: t.RawValue()}
	}
}

// localName returns the fragment/last path segment of an IRI.
func localName(uri string) string {
	if i := strings.LastIndexAny(uri, "#/"); i >= 0 {
		return uri[i+1:]
	}
	return uri
}
