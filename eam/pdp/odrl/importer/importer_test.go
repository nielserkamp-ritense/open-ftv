package importer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

const (
	regoPath = "../../../../testdata/policies/opa/brp/burgerzaken.rego"
	regoSHA  = "e1482e92c4cf0faa3c88bfdc54bc39912fba6d5f6e8db298bc364ac7d5e8bfa8"
)

func newPAP(t *testing.T) pap2.PAP {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return pap2.New(context.Background(), logger, pap2.WithLanguage("odrl"))
}

// odrlDoc renders an AP-NL agreement with one conformsToPolicy constraint and
// a Rego artifact hosted at the given download URL.
func odrlDoc(downloadURL, sha string) string {
	return fmt.Sprintf(`@prefix apnl: <https://standaarden.overheid.nl/odrl-ap-nl/> .
@prefix odrl: <http://www.w3.org/ns/odrl/2/> .
@prefix dct:  <http://purl.org/dc/terms/> .
@prefix dcat: <http://www.w3.org/ns/dcat#> .
@prefix ex:   <https://example.gov.nl/brp/> .

ex:artefact-burgerzaken a apnl:RegoModule ;
    dct:title "Doelbindingsregels burgerzaken (BRP)"@nl ;
    dct:format "application/vnd.rego" ;
    dcat:downloadURL <%s> ;
    apnl:entrypoint "data.doelbinding.burgerzaken.allow" ;
    apnl:sha256 "%s" .

ex:overeenkomst a odrl:Agreement ;
    dct:title "BRP-overeenkomst burgerzaken"@nl ;
    dct:publisher <https://identifier.overheid.nl/tooi/id/oorg/oorg10103> ;
    odrl:profile apnl: ;
    odrl:uid ex:overeenkomst ;
    odrl:assigner <https://identifier.overheid.nl/tooi/id/oorg/oorg10103> ;
    odrl:assignee <https://identifier.overheid.nl/tooi/id/gemeente/gm0014> ;
    odrl:permission [ a odrl:Permission ;
        odrl:action odrl:read ;
        odrl:constraint [ a odrl:Constraint ;
            odrl:leftOperand apnl:verwerkingsverzoek ;
            odrl:operator apnl:conformsToPolicy ;
            odrl:rightOperand ex:artefact-burgerzaken ] ] .
`, downloadURL, sha)
}

func regoServer(t *testing.T) *httptest.Server {
	t.Helper()

	content, err := os.ReadFile(regoPath)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.rego")
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestImportWithArtifactDownload(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	p := newPAP(t)
	imp := New(p, nil, srv.Client())

	doc := odrlDoc(srv.URL+"/burgerzaken.rego", regoSHA)

	res, err := imp.Import(context.Background(), []byte(doc), mime.MimeTypeTurtle)
	require.NoError(t, err)

	require.Len(t, res.Policies, 1)
	assert.Equal(t, "odrl/example.gov.nl-brp-overeenkomst", res.Policies[0])
	require.Len(t, res.Artifacts, 1)
	assert.Equal(t, "rego/doelbinding/burgerzaken", res.Artifacts[0])
	assert.Empty(t, res.Warnings)

	// the ODRL policy is stored as an envelope: raw source + parsed model.
	stored, _, err2 := p.Read("odrl", "example.gov.nl-brp-overeenkomst")
	require.NoError(t, err2)

	data, err3 := io.ReadAll(stored.Content())
	require.NoError(t, err3)

	env, err4 := model.ParseEnvelope(data)
	require.NoError(t, err4)
	assert.Equal(t, "https://example.gov.nl/brp/overeenkomst", env.UID)
	require.Len(t, env.Document.Policies, 1)
	assert.Equal(t, "https://example.gov.nl/brp/overeenkomst", env.Document.Policies[0].UID)

	src, err5 := env.RawSource()
	require.NoError(t, err5)
	assert.Equal(t, doc, string(src))

	// the executable artifact is stored under its own language, ready for the PDP.
	artifact, _, err6 := p.Read("rego", "doelbinding/burgerzaken")
	require.NoError(t, err6)

	content, err7 := io.ReadAll(artifact.Content())
	require.NoError(t, err7)
	want, _ := os.ReadFile(regoPath)
	assert.Equal(t, want, content)
}

func TestImportIsIdempotentOnUID(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	p := newPAP(t)
	imp := New(p, nil, srv.Client())

	doc := []byte(odrlDoc(srv.URL+"/burgerzaken.rego", regoSHA))

	_, err := imp.Import(context.Background(), doc, mime.MimeTypeTurtle)
	require.NoError(t, err)
	_, err = imp.Import(context.Background(), doc, mime.MimeTypeTurtle)
	require.NoError(t, err, "second import of the same uid must not fail")

	list, err2 := p.List("odrl")
	require.NoError(t, err2)
	assert.Len(t, list, 1, "re-import must not create duplicates")
}

func TestImportSha256Mismatch(t *testing.T) {
	t.Parallel()

	srv := regoServer(t)
	p := newPAP(t)
	imp := New(p, nil, srv.Client())

	doc := odrlDoc(srv.URL+"/burgerzaken.rego", "deadbeef"+regoSHA[8:])

	res, err := imp.Import(context.Background(), []byte(doc), mime.MimeTypeTurtle)
	require.NoError(t, err, "the ODRL description itself is still imported")

	assert.Len(t, res.Policies, 1)
	assert.Empty(t, res.Artifacts, "artifact with wrong hash must not be stored")
	require.Len(t, res.Warnings, 1)
	assert.Contains(t, res.Warnings[0], "sha256 mismatch")

	_, _, err2 := p.Read("rego", "doelbinding/burgerzaken")
	assert.Error(t, err2)
}

func TestImportRejectsInvalidDocuments(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	imp := New(p, nil, nil)

	// not parseable.
	_, err := imp.Import(context.Background(), []byte("this is not turtle @@"), mime.MimeTypeTurtle)
	assert.Error(t, err)

	// parseable, but missing the mandatory AP-NL base elements.
	minimal := `@prefix odrl: <http://www.w3.org/ns/odrl/2/> .
<https://example.gov.nl/p> a odrl:Set .`
	_, err = imp.Import(context.Background(), []byte(minimal), mime.MimeTypeTurtle)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not valid ODRL-AP-NL")
}

func TestImportURL(t *testing.T) {
	t.Parallel()

	regoSrv := regoServer(t)
	doc := odrlDoc(regoSrv.URL+"/burgerzaken.rego", regoSHA)

	docSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mime.MimeTypeTurtle)
		_, _ = w.Write([]byte(doc))
	}))
	t.Cleanup(docSrv.Close)

	p := newPAP(t)
	imp := New(p, nil, docSrv.Client())

	res, err := imp.ImportURL(context.Background(), docSrv.URL+"/policy.ttl")
	require.NoError(t, err)
	assert.Len(t, res.Policies, 1)
	assert.Len(t, res.Artifacts, 1)
}

// geoPolicyTTL is example 1 from the ODRL-Geo-NL profile
// (simulaties-2026/bro/odrl-geo-profiel/examples/1-grondwatermonitoring-gevoeligheid.ttl):
// a real geo policy whose odrl:profile set is "apnl: , geonl:".
const geoPolicyTTL = `@prefix geonl:  <https://standaarden.overheid.nl/odrl-geo-nl/> .
@prefix apnl:   <https://standaarden.overheid.nl/odrl-ap-nl/> .
@prefix odrl:   <http://www.w3.org/ns/odrl/2/> .
@prefix xsd:    <http://www.w3.org/2001/XMLSchema#> .
@prefix rdf:    <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs:   <http://www.w3.org/2000/01/rdf-schema#> .
@prefix dct:    <http://purl.org/dc/terms/> .
@prefix dcat:   <http://www.w3.org/ns/dcat#> .
@prefix prov:   <http://www.w3.org/ns/prov#> .
@prefix dpv:    <https://w3id.org/dpv#> .
@prefix prov6:  <https://identifier.overheid.nl/tooi/id/provincie/> .
@prefix oorg:   <https://identifier.overheid.nl/tooi/id/oorg/> .
@prefix bwb:    <https://wetten.overheid.nl/> .
@prefix ex:     <https://example.gov.nl/bro/gmw/> .

ex:ogcApi a dcat:Distribution ;
    dct:title "OGC API Features — GMW"@nl ;
    dcat:accessURL <https://api.example.gov.nl/bro/gmw/v1/collections/putten/items> .

ex:alleenLageGevoeligheid a odrl:Constraint ;
    rdfs:label "Alleen features met gevoeligheid < 4"@nl ;
    odrl:leftOperand geonl:featureProperty ;
    geonl:property <https://example.gov.nl/bro/gmw/def/gevoeligheid> ;
    odrl:operator odrl:lt ;
    odrl:rightOperand "4"^^xsd:integer ;
    apnl:waivable true .

ex:aanbod a odrl:Offer ;
    dct:title "Aanbod BRO-GMW meetnet"@nl ;
    dct:publisher prov6:pv28 ;
    dct:issued "2026-03-01"^^xsd:date ;
    odrl:profile apnl: , geonl: ;
    odrl:uid ex:aanbod ;
    geonl:defaultCRS <http://www.opengis.net/def/crs/EPSG/0/28992> ;
    prov:wasDerivedFrom bwb:BWBR0037095 ;
    odrl:assigner prov6:pv28 ;
    odrl:permission [ a odrl:Permission ;
        odrl:action odrl:read ;
        odrl:target ex:ogcApi ;
        odrl:constraint ex:alleenLageGevoeligheid ] .

ex:overeenkomst-waterschap a odrl:Agreement ;
    dct:title "GMW-overeenkomst Waterschap (overheidsafnemer)"@nl ;
    dct:publisher prov6:pv28 ;
    dct:issued "2026-03-10"^^xsd:date ;
    odrl:profile apnl: , geonl: ;
    odrl:uid ex:overeenkomst-waterschap ;
    apnl:instantiates ex:aanbod ;
    prov:wasDerivedFrom ex:aanbod ;
    odrl:assigner prov6:pv28 ;
    odrl:assignee oorg:oorg41000 ;
    odrl:permission [ a odrl:Permission ;
        odrl:action odrl:read ;
        odrl:target ex:ogcApi ] .
`

// TestImportGeoProfilePolicy verifies that the importer accepts a real
// ODRL-Geo-NL policy whose odrl:profile set contains the ODRL-AP-NL base
// profile alongside the geonl profile (regression for the geonl rejection).
func TestImportGeoProfilePolicy(t *testing.T) {
	t.Parallel()

	p := newPAP(t)
	imp := New(p, nil, nil)

	res, err := imp.Import(context.Background(), []byte(geoPolicyTTL), mime.MimeTypeTurtle)
	require.NoError(t, err, "a policy carrying apnl + geonl profiles must be accepted")

	require.Len(t, res.Policies, 2)
	// geonl is a known extra profile, so no warning is emitted.
	assert.Empty(t, res.Warnings)

	// the stored envelope preserves the full odrl:profile set.
	stored, _, err2 := p.Read("odrl", SanitizeUID("https://example.gov.nl/bro/gmw/aanbod"))
	require.NoError(t, err2)
	data, err3 := io.ReadAll(stored.Content())
	require.NoError(t, err3)
	env, err4 := model.ParseEnvelope(data)
	require.NoError(t, err4)
	require.Len(t, env.Document.Policies, 1)
	pol := env.Document.Policies[0]
	assert.Equal(t, model.APNLProfile, pol.Profile)
	assert.Contains(t, pol.Profiles, model.APNLProfile)
	assert.Contains(t, pol.Profiles, model.GeoNLProfile)
}

// TestValidateUnknownExtraProfileWarns verifies that an unrecognised extra
// profile is accepted with a warning under the default options, and rejected
// when RejectUnknownProfiles is set.
func TestValidateUnknownExtraProfileWarns(t *testing.T) {
	t.Parallel()

	pol := &model.Policy{
		UID:       "https://example.gov.nl/p",
		Titles:    []model.LangString{{Value: "t"}},
		Publisher: "https://example.gov.nl/pub",
		Profiles:  []string{model.APNLProfile, "https://example.gov.nl/some-unknown-profile/"},
	}

	warnings, err := model.ValidatePolicyOpts(pol, model.ValidateOptions{})
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "unrecognised profile")

	_, err = model.ValidatePolicyOpts(pol, model.ValidateOptions{RejectUnknownProfiles: true})
	assert.Error(t, err)
}

func TestSanitizeUID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "example.gov.nl-brf-brfAanbod", SanitizeUID("https://example.gov.nl/brf/brfAanbod"))
	assert.Equal(t, "ftv-odrl-x", SanitizeUID("urn:ftv:odrl:x"))
}

func TestArtifactID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "doelbinding/burgerzaken", ArtifactID(&model.Artifact{
		UID:        "https://x.nl/a",
		Entrypoint: "data.doelbinding.burgerzaken.allow",
	}))
	assert.Equal(t, "doelbinding", ArtifactID(&model.Artifact{
		UID:        "https://x.nl/a",
		Entrypoint: "data.doelbinding",
	}))
	assert.Equal(t, "x.nl-a", ArtifactID(&model.Artifact{UID: "https://x.nl/a"}))
}
