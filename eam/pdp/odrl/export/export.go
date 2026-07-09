package export

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/rdf"
)

// Exporter renders the PAP-administered policies as an ODRL-AP-NL document.
type Exporter struct {
	pap     pap2.PAP
	ann     *Annotations
	baseURL string
}

// New creates an Exporter for the given PAP.
//
// baseURL is the public base URL of this PAP/manager (e.g.
// "https://pap.example.gov.nl"); it is used for default artifact
// download-URLs and generated uids. ann may be nil.
func New(p pap2.PAP, ann *Annotations, baseURL string) *Exporter {
	if ann == nil {
		ann = &Annotations{}
	}
	return &Exporter{pap: p, ann: ann, baseURL: strings.TrimSuffix(baseURL, "/")}
}

// ExportAll describes every exportable policy in the PAP as one ODRL-AP-NL
// document: executable policies become PolicyArtifacts under a single
// Offer/Set; policies stored as language "odrl" (previously imported
// documents) are merged in as-is.
func (e *Exporter) ExportAll() (*model.Document, error) {
	list, err := e.pap.List("")
	if err != nil {
		return nil, err
	}

	doc := &model.Document{}
	set := e.newSetPolicy()

	for i := range list {
		pol := list[i]

		if strings.EqualFold(pol.Language(), "odrl") {
			e.mergeODRL(doc, pol)
			continue
		}

		artifact, perm, err2 := e.describe(pol)
		if err2 != nil || artifact == nil {
			continue // language without an AP-NL artifact mapping.
		}

		doc.Artifacts = append(doc.Artifacts, artifact)
		set.Permissions = append(set.Permissions, perm)
	}

	if len(set.Permissions) > 0 {
		doc.Policies = append(doc.Policies, set)
	}
	return doc, nil
}

// ExportPolicy describes a single policy key as an ODRL-AP-NL document.
func (e *Exporter) ExportPolicy(language, id string) (*model.Document, error) {
	pol, _, err := e.pap.Read(language, id)
	if err != nil {
		return nil, err
	}

	doc := &model.Document{}

	if strings.EqualFold(pol.Language(), "odrl") {
		e.mergeODRL(doc, pol)
		return doc, nil
	}

	artifact, perm, err2 := e.describe(pol)
	if err2 != nil {
		return nil, err2
	}
	if artifact == nil {
		return nil, fmt.Errorf("policy language %q has no ODRL-AP-NL artifact mapping", pol.Language())
	}

	set := e.newSetPolicy()
	set.UID = e.uidBase() + "policy/" + pol.Key()
	set.Permissions = []*model.Rule{perm}

	doc.Artifacts = append(doc.Artifacts, artifact)
	doc.Policies = append(doc.Policies, set)
	return doc, nil
}

// describe builds the PolicyArtifact and the permission that references it for
// one executable policy. It returns (nil, nil, nil) when the policy language
// has no artifact mapping.
func (e *Exporter) describe(pol pap2.Policy) (*model.Artifact, *model.Rule, error) {
	at, ok := model.LanguageArtifact(strings.ToLower(pol.Language()))
	if !ok {
		return nil, nil, nil
	}

	content, err := io.ReadAll(pol.Content())
	if err != nil {
		return nil, nil, err
	}

	key := pol.Key()
	meta := e.ann.Policies[key]
	sum := sha256.Sum256(content)

	artifact := &model.Artifact{
		UID:         e.uidBase() + "artifact/" + key,
		Type:        at.Class,
		Format:      at.MediaType,
		SHA256:      hex.EncodeToString(sum[:]),
		DownloadURL: e.downloadURL(pol, meta),
		Entrypoint:  e.entrypoint(pol, meta, content),
	}
	if meta.Title != "" {
		artifact.Titles = []model.LangString{{Value: meta.Title, Language: "nl"}}
	}
	if meta.Description != "" {
		artifact.Descriptions = []model.LangString{{Value: meta.Description, Language: "nl"}}
	}

	action := meta.Action
	if action == "" {
		action = rdf.ODRLUse
	}

	perm := &model.Rule{
		Action: &model.Action{Value: action},
		Constraints: []*model.Constraint{{
			LeftOperand:  model.APNLVerwerkingsverzoek,
			Operator:     model.APNLConformsToPolicy,
			RightOperand: []model.Value{{ID: artifact.UID}},
		}},
	}

	if meta.Purpose != "" {
		perm.Action.Refinements = []*model.Constraint{{
			LeftOperand:  rdf.URIODRL + "purpose",
			Operator:     rdf.ODRLIsA,
			RightOperand: []model.Value{{ID: meta.Purpose}},
		}}
	}

	if target := firstNonEmpty(meta.Target, e.ann.Policy.Target); target != "" {
		perm.Targets = []string{target}
	}

	// Requested BRP field-groups (rubrieken): emit each as brp:verzochteRubriek
	// (a subPropertyOf odrl:target) AND as a plain odrl:target, so both
	// profile-aware harvesters and generic ODRL tooling see the fields (C1).
	if len(meta.Velden) > 0 {
		var rubrieken []model.Value
		for _, veld := range meta.Velden {
			if veld == "" {
				continue
			}
			perm.Targets = append(perm.Targets, veld)
			rubrieken = append(rubrieken, model.Value{ID: veld})
		}
		if len(rubrieken) > 0 {
			perm.Extra = append(perm.Extra, model.Property{
				Predicate: model.BRPVerzochteRubriek,
				Values:    rubrieken,
			})
		}
	}

	// Record-bound condition rules (voorwaarderegels): emit each as an
	// odrl:constraint on the permission in the exact blank-node shape the
	// BRP-sim harvester recognises, so an export-harvested profile keeps the
	// per-record rules (e.g. WOZ brp:knv on overlijden.datum) instead of losing
	// them.
	for _, vw := range meta.Voorwaarden {
		if c := voorwaardeConstraint(vw); c != nil {
			perm.Constraints = append(perm.Constraints, c)
		}
	}

	return artifact, perm, nil
}

// voorwaardeConstraint builds the odrl:constraint for one condition rule,
// resolving short leftOperand/operator names to their BRP IRIs. It returns nil
// when the rule has no left operand and operator to constrain on.
func voorwaardeConstraint(vw VoorwaardeMeta) *model.Constraint {
	left := resolveIRI(vw.LeftOperand, model.BRPRubriekNamespace)
	operator := resolveIRI(vw.Operator, model.BRPDefNamespace)
	if left == "" || operator == "" {
		return nil
	}
	c := &model.Constraint{LeftOperand: left, Operator: operator}
	if vw.RightOperand != "" {
		if isIRI(vw.RightOperand) {
			c.RightOperand = []model.Value{{ID: vw.RightOperand}}
		} else {
			c.RightOperand = []model.Value{{Value: vw.RightOperand}}
		}
	}
	return c
}

// resolveIRI returns value unchanged when it already is an IRI, otherwise
// prefixes the bare short name with ns.
func resolveIRI(value, ns string) string {
	if value == "" {
		return ""
	}
	if isIRI(value) {
		return value
	}
	return ns + value
}

// isIRI reports whether s looks like an absolute IRI rather than a short name.
func isIRI(s string) bool {
	return strings.Contains(s, "://") || strings.HasPrefix(s, "urn:")
}

// mergeODRL merges a previously imported ODRL policy (stored as an envelope)
// into the output document.
func (e *Exporter) mergeODRL(doc *model.Document, pol pap2.Policy) {
	data, err := io.ReadAll(pol.Content())
	if err != nil {
		return
	}

	env, err2 := model.ParseEnvelope(data)
	if err2 != nil || env.Document == nil {
		return
	}

	for _, p := range env.Document.Policies {
		if doc.Policy(p.UID) == nil {
			doc.Policies = append(doc.Policies, p)
		}
	}
	for _, a := range env.Document.Artifacts {
		if doc.Artifact(a.UID) == nil {
			doc.Artifacts = append(doc.Artifacts, a)
		}
	}
	doc.Bundles = append(doc.Bundles, env.Document.Bundles...)
	doc.Constraints = append(doc.Constraints, env.Document.Constraints...)
	doc.Duties = append(doc.Duties, env.Document.Duties...)
}

func (e *Exporter) newSetPolicy() *model.Policy {
	m := e.ann.Policy

	typ := rdf.ODRLOffer
	switch strings.ToLower(m.Type) {
	case "set":
		typ = rdf.ODRLSet
	case "agreement":
		typ = rdf.ODRLAgreement
	}

	uid := m.UID
	if uid == "" {
		uid = e.uidBase() + "policy"
	}

	title := m.Title
	if title == "" {
		title = "Machine-uitvoerbaar toegangsbeleid (PAP-export)"
	}

	p := &model.Policy{
		UID:          uid,
		Type:         typ,
		Profile:      model.APNLProfile,
		Titles:       []model.LangString{{Value: title, Language: "nl"}},
		Publisher:    m.Publisher,
		Assigner:     m.Assigner,
		Assignee:     m.Assignee,
		Issued:       m.Issued,
		Instantiates: m.Instantiates,
	}
	if m.Description != "" {
		p.Descriptions = []model.LangString{{Value: m.Description, Language: "nl"}}
	}
	return p
}

func (e *Exporter) uidBase() string {
	if e.ann.UIDBase != "" {
		return e.ann.UIDBase
	}
	if e.baseURL != "" {
		return e.baseURL + "/odrl/"
	}
	return "urn:ftv:odrl:"
}

// downloadURL is the location where the artifact content can be retrieved:
// the policy's own url when registered, an annotation override, or the
// policy endpoint of this PAP.
func (e *Exporter) downloadURL(pol pap2.Policy, meta ArtifactMeta) string {
	if meta.DownloadURL != "" {
		return meta.DownloadURL
	}
	if pol.URI() != "" {
		return pol.URI()
	}
	if e.baseURL != "" {
		return e.baseURL + "/v1/policy/" + pol.Key()
	}
	return ""
}

var regoPackage = regexp.MustCompile(`(?m)^\s*package\s+([a-zA-Z0-9_.\[\]"]+)`)

// entrypoint determines the apnl:entrypoint: an annotation override, the Rego
// package ("data.<package>") or, as a fallback, the policy key.
func (e *Exporter) entrypoint(pol pap2.Policy, meta ArtifactMeta, content []byte) string {
	if meta.Entrypoint != "" {
		return meta.Entrypoint
	}

	lang := strings.ToLower(pol.Language())
	if lang == "opa" || lang == "rego" || strings.HasPrefix(lang, "opa/") {
		if m := regoPackage.FindSubmatch(content); m != nil {
			return "data." + string(m[1])
		}
	}

	return pol.Key()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
