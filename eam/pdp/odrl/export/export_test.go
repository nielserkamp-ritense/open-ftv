package export

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

const policies = "../../../../testdata/policies"

func newPAP(t *testing.T) pap2.PAP {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return pap2.New(context.Background(), logger, pap2.WithLanguage("rego"))
}

func addPolicy(t *testing.T, p pap2.PAP, language, id, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	pol, err2 := pap2.NewPolicyFromData(id, language, "", "", bytes.NewReader(content))
	require.NoError(t, err2)

	_, err3 := p.Create(pol)
	require.NoError(t, err3)
	return content
}

func TestExportRealRegoPolicy(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	content := addPolicy(t, p, "opa", "dienst_toeslagen/zorgtoeslag",
		filepath.Join(policies, "opa/dienst_toeslagen/zorgtoeslag.rego"))

	sum := sha256.Sum256(content)
	wantSHA := hex.EncodeToString(sum[:])

	ann := &Annotations{
		UIDBase: "https://pap.example.gov.nl/odrl/",
		Policy: PolicyMeta{
			Type:      "Offer",
			Title:     "Aanbod toeslagen-beleid",
			Publisher: "https://identifier.overheid.nl/tooi/id/oorg/oorg10103",
			Assigner:  "https://identifier.overheid.nl/tooi/id/oorg/oorg10103",
			Issued:    "2026-07-08",
			Target:    "https://data.example.gov.nl/toeslagen/distributie",
		},
		Policies: map[string]ArtifactMeta{
			"opa/dienst_toeslagen/zorgtoeslag": {
				Title:   "Doelbindingsregels zorgtoeslag",
				Purpose: "https://register.example.gov.nl/verwerkingen/zorgtoeslag",
			},
		},
	}

	e := New(p, ann, "https://pap.example.gov.nl")

	doc, err := e.ExportAll()
	require.NoError(t, err)

	require.Len(t, doc.Artifacts, 1)
	a := doc.Artifacts[0]
	assert.Equal(t, "https://pap.example.gov.nl/odrl/artifact/opa/dienst_toeslagen/zorgtoeslag", a.UID)
	assert.Equal(t, model.APNLRegoModule, a.Type)
	assert.Equal(t, "application/vnd.rego", a.Format)
	assert.Equal(t, wantSHA, a.SHA256)
	assert.Equal(t, "data.doelbinding.zorgtoeslag", a.Entrypoint)
	assert.Equal(t, "https://pap.example.gov.nl/v1/policy/opa/dienst_toeslagen/zorgtoeslag", a.DownloadURL)

	require.Len(t, doc.Policies, 1)
	pol := doc.Policies[0]
	assert.Equal(t, "http://www.w3.org/ns/odrl/2/Offer", pol.Type)
	assert.Equal(t, model.APNLProfile, pol.Profile)
	assert.NoError(t, model.ValidatePolicy(pol))
	require.Len(t, pol.Permissions, 1)

	perm := pol.Permissions[0]
	require.Len(t, perm.Constraints, 1)
	c := perm.Constraints[0]
	assert.Equal(t, model.APNLVerwerkingsverzoek, c.LeftOperand)
	assert.Equal(t, model.APNLConformsToPolicy, c.Operator)
	require.Len(t, c.RightOperand, 1)
	assert.Equal(t, a.UID, c.RightOperand[0].ID)

	require.NotNil(t, perm.Action)
	require.Len(t, perm.Action.Refinements, 1)
	assert.Equal(t, "https://register.example.gov.nl/verwerkingen/zorgtoeslag",
		perm.Action.Refinements[0].RightOperand[0].ID)
	assert.Equal(t, []string{"https://data.example.gov.nl/toeslagen/distributie"}, perm.Targets)

	// the export must serialize to valid Turtle and JSON-LD and parse back.
	for _, mt := range []string{mime.MimeTypeTurtle, mime.MimeTypeJSONLD} {
		var buf bytes.Buffer
		require.NoError(t, doc.Serialize(&buf, mt))

		doc2, err2 := model.Parse(&buf, mt)
		require.NoError(t, err2)
		require.Len(t, doc2.Artifacts, 1)
		assert.Equal(t, wantSHA, doc2.Artifacts[0].SHA256)
		require.Len(t, doc2.Policies, 1)
	}
}

// TestExportVeldenAsVerzochteRubriek covers the C1 export side: an annotation
// with per-policy velden must be emitted as brp:verzochteRubriek AND as
// odrl:target, so the Turtle carries the requested rubrieken.
func TestExportVeldenAsVerzochteRubriek(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	addPolicy(t, p, "opa", "brp/burgerzaken", filepath.Join(policies, "opa/brp/burgerzaken.rego"))

	rubriek1 := "https://data.rijksoverheid.nl/brp/rubriek/naam"
	rubriek2 := "https://data.rijksoverheid.nl/brp/rubriek/geboorte"

	ann := &Annotations{
		UIDBase: "https://pap.example.gov.nl/odrl/",
		Policy:  PolicyMeta{Type: "Offer", Title: "Aanbod BRP"},
		Policies: map[string]ArtifactMeta{
			"opa/brp/burgerzaken": {
				Title:  "Doelbindingsregels burgerzaken",
				Target: "https://data.example.gov.nl/brp/distributie",
				Velden: []string{rubriek1, rubriek2},
			},
		},
	}

	e := New(p, ann, "https://pap.example.gov.nl")

	doc, err := e.ExportAll()
	require.NoError(t, err)
	require.Len(t, doc.Policies, 1)
	perm := doc.Policies[0].Permissions[0]

	// both rubrieken present as odrl:target (alongside the distribution target).
	assert.Subset(t, perm.Targets, []string{rubriek1, rubriek2})
	// and as brp:verzochteRubriek in Extra.
	var found []string
	for _, ex := range perm.Extra {
		if ex.Predicate == model.BRPVerzochteRubriek {
			for _, v := range ex.Values {
				found = append(found, v.ID)
			}
		}
	}
	assert.ElementsMatch(t, []string{rubriek1, rubriek2}, found)

	// the serialized Turtle must contain the verzochteRubriek triples.
	var buf bytes.Buffer
	require.NoError(t, doc.Serialize(&buf, mime.MimeTypeTurtle))
	ttl := buf.String()
	assert.Contains(t, ttl, model.BRPVerzochteRubriek)
	assert.Contains(t, ttl, rubriek1)
	assert.Contains(t, ttl, rubriek2)
}

// TestExportVoorwaardenAsConstraint covers the C3 export side: a policy
// annotation with voorwaarden must be emitted as odrl:constraint on the
// permission (leftOperand rubriek-IRI, operator brp-def-IRI, optional
// rightOperand), and must survive a Turtle round-trip through the model parser
// in the exact shape the BRP-sim harvester recognises.
func TestExportVoorwaardenAsConstraint(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	addPolicy(t, p, "opa", "brp/burgerzaken", filepath.Join(policies, "opa/brp/burgerzaken.rego"))

	overlijden := "https://data.rijksoverheid.nl/brp/rubriek/datumOverlijdenOverlijden"
	knv := "https://data.rijksoverheid.nl/brp/def#knv"
	gemeenteIRI := "https://identifier.overheid.nl/tooi/id/gemeente/gm0014"

	ann := &Annotations{
		UIDBase: "https://pap.example.gov.nl/odrl/",
		Policy:  PolicyMeta{Type: "Offer", Title: "Aanbod BRP"},
		Policies: map[string]ArtifactMeta{
			"opa/brp/burgerzaken": {
				Title: "Doelbindingsregels WOZ",
				Voorwaarden: []VoorwaardeMeta{
					// short names must resolve to the brp rubriek/def IRIs.
					{LeftOperand: "datumOverlijdenOverlijden", Operator: "knv"},
					// full IRIs and a rightOperand IRI must pass through unchanged.
					{LeftOperand: "https://data.rijksoverheid.nl/brp/rubriek/gemeenteVanInschrijvingVerblijfplaats",
						Operator: "ga1", RightOperand: gemeenteIRI},
				},
			},
		},
	}

	e := New(p, ann, "https://pap.example.gov.nl")

	doc, err := e.ExportAll()
	require.NoError(t, err)
	require.Len(t, doc.Policies, 1)
	perm := doc.Policies[0].Permissions[0]

	// find the knv voorwaarde constraint (alongside the conformsToPolicy one).
	knvC := findConstraint(perm.Constraints, overlijden, knv)
	require.NotNil(t, knvC, "knv constraint must be emitted")
	assert.Empty(t, knvC.RightOperand)

	ga1C := findConstraint(perm.Constraints,
		"https://data.rijksoverheid.nl/brp/rubriek/gemeenteVanInschrijvingVerblijfplaats",
		"https://data.rijksoverheid.nl/brp/def#ga1")
	require.NotNil(t, ga1C, "ga1 constraint must be emitted")
	require.Len(t, ga1C.RightOperand, 1)
	assert.Equal(t, gemeenteIRI, ga1C.RightOperand[0].ID)

	// the Turtle must carry the constraint, and it must round-trip through the
	// model parser (same predicates/IRIs the BRP-sim parser reads).
	var buf bytes.Buffer
	require.NoError(t, doc.Serialize(&buf, mime.MimeTypeTurtle))
	ttl := buf.String()
	assert.Contains(t, ttl, overlijden)
	assert.Contains(t, ttl, knv)

	raw := buf.Bytes()
	doc2, err2 := model.Parse(bytes.NewReader(raw), mime.MimeTypeTurtle)
	require.NoError(t, err2)
	require.Len(t, doc2.Policies, 1)
	perm2 := doc2.Policies[0].Permissions[0]
	assert.NotNil(t, findConstraint(perm2.Constraints, overlijden, knv),
		"knv constraint must survive the round-trip")
	rt := findConstraint(perm2.Constraints,
		"https://data.rijksoverheid.nl/brp/rubriek/gemeenteVanInschrijvingVerblijfplaats",
		"https://data.rijksoverheid.nl/brp/def#ga1")
	require.NotNil(t, rt)
	require.Len(t, rt.RightOperand, 1)
	assert.Equal(t, gemeenteIRI, rt.RightOperand[0].ID)
}

// findConstraint returns the constraint matching leftOperand and operator, or nil.
func findConstraint(cs []*model.Constraint, left, op string) *model.Constraint {
	for _, c := range cs {
		if c.LeftOperand == left && c.Operator == op {
			return c
		}
	}
	return nil
}

func TestExportPolicySingleKey(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	addPolicy(t, p, "opa", "brp/burgerzaken", filepath.Join(policies, "opa/brp/burgerzaken.rego"))

	e := New(p, nil, "https://pap.example.gov.nl")

	doc, err := e.ExportPolicy("opa", "brp/burgerzaken")
	require.NoError(t, err)
	require.Len(t, doc.Artifacts, 1)

	// example 3 of the profile documents this exact hash for burgerzaken.rego.
	assert.Equal(t, "e1482e92c4cf0faa3c88bfdc54bc39912fba6d5f6e8db298bc364ac7d5e8bfa8", doc.Artifacts[0].SHA256)
	assert.Equal(t, "data.doelbinding.burgerzaken", doc.Artifacts[0].Entrypoint)

	_, err2 := e.ExportPolicy("opa", "does/not/exist")
	assert.Error(t, err2)
}

func TestExportSkipsUnknownLanguages(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	pol, err := pap2.NewPolicyFromData("some-policy", "xacml", "", "", bytes.NewReader([]byte("<xml/>")))
	require.NoError(t, err)
	_, err = p.Create(pol)
	require.NoError(t, err)

	e := New(p, nil, "")
	doc, err2 := e.ExportAll()
	require.NoError(t, err2)
	assert.Empty(t, doc.Artifacts)
	assert.Empty(t, doc.Policies)
}

func TestLoadAnnotations(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, AnnotationsFile)
	require.NoError(t, os.WriteFile(path, []byte(`
uidBase: "https://x.nl/odrl/"
policy:
  type: Set
  title: "Test"
policies:
  opa/a/b:
    purpose: "https://x.nl/doel"
`), 0o644))

	a, err := LoadAnnotations(path)
	require.NoError(t, err)
	assert.Equal(t, "https://x.nl/odrl/", a.UIDBase)
	assert.Equal(t, "Set", a.Policy.Type)
	assert.Equal(t, "https://x.nl/doel", a.Policies["opa/a/b"].Purpose)

	// missing file is not an error.
	a2, err2 := LoadAnnotations(filepath.Join(dir, "missing.yaml"))
	require.NoError(t, err2)
	assert.Empty(t, a2.UIDBase)
}
