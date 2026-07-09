package model

import (
	"io"
	"strconv"

	"github.com/deiu/rdf2go"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/rdf"
)

// Serialize writes the document to w in the given mime type
// ("text/turtle" or "application/ld+json").
func (d *Document) Serialize(w io.Writer, mimeType string) error {
	g := d.Graph()
	if mimeType == mime.MimeTypeJSONLD {
		return serializeJSONLD(g, w)
	}
	return g.Serialize(w, mime.MimeTypeTurtle)
}

// Graph builds an RDF graph from the document.
func (d *Document) Graph() *rdf2go.Graph {
	b := &builder{g: rdf2go.NewGraph("")}
	for _, c := range d.Constraints {
		b.constraint(c)
	}
	for _, du := range d.Duties {
		b.duty(du)
	}
	for _, a := range d.Artifacts {
		b.artifact(a)
	}
	for _, bu := range d.Bundles {
		b.bundle(bu)
	}
	for _, p := range d.Policies {
		b.policy(p)
	}
	return b.g
}

type builder struct {
	g   *rdf2go.Graph
	bnc int
}

func (b *builder) blank() rdf2go.Term {
	b.bnc++
	return rdf2go.NewBlankNode("b" + strconv.Itoa(b.bnc))
}

func (b *builder) addRes(s rdf2go.Term, p, o string) {
	if o == "" {
		return
	}
	b.g.AddTriple(s, res(p), res(o))
}

func (b *builder) addLit(s rdf2go.Term, p, o string) {
	if o == "" {
		return
	}
	b.g.AddTriple(s, res(p), rdf2go.NewLiteral(o))
}

func (b *builder) addTyped(s rdf2go.Term, p, o, dt string) {
	if o == "" {
		return
	}
	b.g.AddTriple(s, res(p), rdf2go.NewLiteralWithDatatype(o, res(dt)))
}

func (b *builder) addLangs(s rdf2go.Term, p string, ls []LangString) {
	for _, l := range ls {
		if l.Language != "" {
			b.g.AddTriple(s, res(p), rdf2go.NewLiteralWithLanguage(l.Value, l.Language))
		} else {
			b.g.AddTriple(s, res(p), rdf2go.NewLiteral(l.Value))
		}
	}
}

func (b *builder) value(s rdf2go.Term, p string, v Value) {
	switch {
	case v.IsRef():
		b.g.AddTriple(s, res(p), res(v.ID))
	case v.Language != "":
		b.g.AddTriple(s, res(p), rdf2go.NewLiteralWithLanguage(v.Value, v.Language))
	case v.Datatype != "":
		b.g.AddTriple(s, res(p), rdf2go.NewLiteralWithDatatype(v.Value, res(v.Datatype)))
	default:
		b.g.AddTriple(s, res(p), rdf2go.NewLiteral(v.Value))
	}
}

func (b *builder) props(s rdf2go.Term, ps []Property) {
	for _, pr := range ps {
		for _, v := range pr.Values {
			b.value(s, pr.Predicate, v)
		}
	}
}

func (b *builder) policy(p *Policy) {
	s := res(p.UID)
	typ := p.Type
	if typ == "" {
		typ = rdf.ODRLSet
	}
	b.addRes(s, rdf.Type, typ)
	b.addRes(s, rdf.ODRLIdentifier, p.UID)
	for _, prof := range p.profileSet() {
		b.addRes(s, rdf.ODRLProfile, prof)
	}
	b.addLangs(s, DCTTitle, p.Titles)
	b.addLangs(s, DCTDescription, p.Descriptions)
	b.addRes(s, DCTPublisher, p.Publisher)
	if p.Issued != "" {
		b.addTyped(s, DCTIssued, p.Issued, XSDDate)
	}
	b.addRes(s, rdf.ODRLAssigner, p.Assigner)
	b.addRes(s, rdf.ODRLAssignee, p.Assignee)
	b.addRes(s, APNLInstantiates, p.Instantiates)
	b.addRes(s, PROVWasDerivedFrom, p.WasDerivedFrom)
	b.addRes(s, PROVWasRevisionOf, p.WasRevisionOf)
	for _, r := range p.Permissions {
		b.g.AddTriple(s, res(rdf.ODRLHasPermission), b.rule(r))
	}
	for _, r := range p.Prohibitions {
		b.g.AddTriple(s, res(rdf.ODRLHasProhibition), b.rule(r))
	}
	for _, du := range p.Obligations {
		b.g.AddTriple(s, res(rdf.ODRLObligation), b.dutyRef(du))
	}
	b.props(s, p.Extra)
}

func (b *builder) rule(r *Rule) rdf2go.Term {
	var s rdf2go.Term
	if r.UID != "" {
		s = res(r.UID)
	} else {
		s = b.blank()
	}
	if r.Action != nil {
		b.g.AddTriple(s, res(rdf.ODRLHasAction), b.action(r.Action))
	}
	for _, t := range r.Targets {
		b.addRes(s, ODRLTargetProp, t)
	}
	b.addRes(s, rdf.ODRLAssigner, r.Assigner)
	b.addRes(s, rdf.ODRLAssignee, r.Assignee)
	for _, c := range r.Constraints {
		b.g.AddTriple(s, res(rdf.ODRLHasConstraint), b.constraintRef(c))
	}
	for _, du := range r.Duties {
		b.g.AddTriple(s, res(rdf.ODRLHasDuty), b.dutyRef(du))
	}
	b.props(s, r.Extra)
	return s
}

func (b *builder) action(a *Action) rdf2go.Term {
	if len(a.Refinements) == 0 && a.InformedParty == "" {
		if a.Value != "" {
			return res(a.Value)
		}
	}
	s := b.blank()
	b.addRes(s, rdf.Value, a.Value)
	for _, c := range a.Refinements {
		b.g.AddTriple(s, res(rdf.ODRLRefinement), b.constraintRef(c))
	}
	b.addRes(s, ODRLInformedParty, a.InformedParty)
	return s
}

func (b *builder) constraintRef(c *Constraint) rdf2go.Term {
	if c.Ref != "" {
		return res(c.Ref)
	}
	return b.constraint(c)
}

func (b *builder) constraint(c *Constraint) rdf2go.Term {
	if c.Ref != "" {
		return res(c.Ref)
	}
	var s rdf2go.Term
	if c.UID != "" {
		s = res(c.UID)
	} else {
		s = b.blank()
	}
	b.addRes(s, rdf.Type, rdf.ODRLConstraint)
	b.addRes(s, rdf.ODRLHasLeftOperand, c.LeftOperand)
	b.addRes(s, rdf.ODRLOperator, c.Operator)
	for _, v := range c.RightOperand {
		b.value(s, rdf.ODRLHasRightOperand, v)
	}
	b.addLangs(s, RDFSLabel, c.Labels)
	b.addLangs(s, RDFSComment, c.Comments)
	if c.Waivable != nil {
		b.addTyped(s, APNLWaivable, strconv.FormatBool(*c.Waivable), XSDBoolean)
	}
	if c.LogicalOp != "" && len(c.Operands) > 0 {
		members := make([]rdf2go.Term, len(c.Operands))
		for i, op := range c.Operands {
			members[i] = b.constraintRef(op)
		}
		b.g.AddTriple(s, res(c.LogicalOp), b.rdfList(members))
	}
	return s
}

func (b *builder) duty(d *Duty) rdf2go.Term {
	var s rdf2go.Term
	if d.UID != "" {
		s = res(d.UID)
	} else {
		s = b.blank()
	}
	b.addRes(s, rdf.Type, rdf.ODRLDuty)
	if d.Action != nil {
		b.g.AddTriple(s, res(rdf.ODRLHasAction), b.action(d.Action))
	}
	b.addLangs(s, RDFSLabel, d.Labels)
	return s
}

func (b *builder) dutyRef(d *Duty) rdf2go.Term {
	if d.Ref != "" {
		return res(d.Ref)
	}
	return b.duty(d)
}

func (b *builder) artifact(a *Artifact) rdf2go.Term {
	s := res(a.UID)
	typ := a.Type
	if typ == "" {
		typ = APNLPolicyArtifact
	}
	b.addRes(s, rdf.Type, typ)
	b.addLangs(s, DCTTitle, a.Titles)
	b.addLangs(s, DCTDescription, a.Descriptions)
	b.addLit(s, DCTFormat, a.Format)
	b.addRes(s, DCATDownloadURL, a.DownloadURL)
	b.addLit(s, APNLEntrypoint, a.Entrypoint)
	b.addLit(s, APNLSha256, a.SHA256)
	b.addRes(s, PROVWasRevisionOf, a.WasRevisionOf)
	b.addRes(s, PROVWasDerivedFrom, a.WasDerivedFrom)
	return s
}

func (b *builder) bundle(bu *Bundle) rdf2go.Term {
	s := res(bu.UID)
	b.addRes(s, rdf.Type, APNLPolicyBundle)
	b.addLangs(s, DCTTitle, bu.Titles)
	b.addLangs(s, DCTDescription, bu.Descriptions)
	b.addRes(s, DCTPublisher, bu.Publisher)
	if bu.Issued != "" {
		b.addTyped(s, DCTIssued, bu.Issued, XSDDate)
	}
	b.addLit(s, DCTFormat, bu.Format)
	b.addRes(s, DCATDownloadURL, bu.DownloadURL)
	b.addLit(s, APNLEntrypoint, bu.Entrypoint)
	b.addLit(s, APNLSha256, bu.SHA256)
	b.addRes(s, PROVWasRevisionOf, bu.WasRevisionOf)
	for _, art := range bu.Bundles {
		b.addRes(s, APNLBundles, art)
	}
	return s
}

func (b *builder) rdfList(members []rdf2go.Term) rdf2go.Term {
	nilRes := res(rdf.URIRDF + "nil")
	if len(members) == 0 {
		return nilRes
	}
	head := b.blank()
	cur := head
	for i, m := range members {
		b.g.AddTriple(cur, res(rdf.URIRDF+"first"), m)
		if i == len(members)-1 {
			b.g.AddTriple(cur, res(rdf.URIRDF+"rest"), nilRes)
		} else {
			next := b.blank()
			b.g.AddTriple(cur, res(rdf.URIRDF+"rest"), next)
			cur = next
		}
	}
	return head
}
