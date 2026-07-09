package odrl_geo

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}

// newController wires a real in-memory PAP holding the example policies (stored
// as "odrl" envelopes, exactly the importer's storage format), an optional PIP,
// and the ODRL-Geo engine. Envelopes are stored directly rather than via the
// importer because the shared ODRL-AP-NL validator (which this module must not
// modify) currently rejects the geonl profile; see the module notes.
func newController(t *testing.T, entities []models.Entity, ttls ...string) pdp.Controller {
	t.Helper()
	ctx := context.Background()
	logger := testLogger()

	p := pap.New(ctx, logger)
	for _, name := range ttls {
		src, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		id := strings.TrimSuffix(name, ".ttl")
		stored, err := pap.NewPolicyFromData(id, "odrl", "", "urn:"+id, bytes.NewReader(wrapEnvelope(t, src)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Create(stored); err != nil {
			t.Fatalf("store %s: %v", name, err)
		}
	}

	ip := pip.New(ctx, logger)
	for _, e := range entities {
		ip.AddEntity(e)
	}

	return NewController(pdp.WithContext(ctx), pdp.WithLogger(logger), pdp.WithPAP(p), pdp.WithPIP(ip))
}

// req builds a PARC for a feature request.
func req(subject, purpose, layer, crs, geometry string, attrs map[string]any) *models.PARC {
	return typedReq(resourceFeature, subject, purpose, layer, crs, geometry, attrs)
}

func typedReq(resType, subject, purpose, layer, crs, geometry string, attrs map[string]any) *models.PARC {
	actionProps := map[string]any{}
	if purpose != "" {
		actionProps[actionProcessingActivity] = purpose
	}
	resProps := map[string]any{propLayer: layer, propCRS: crs}
	if geometry != "" {
		resProps[propGeometry] = geometry
	}
	if attrs != nil {
		resProps[propAttributes] = attrs
	}
	return &models.PARC{
		Principal: models.NewEntity("organization", subject, models.NewAttributeSet(nil)),
		Action:    models.NewEntity("name", "read", models.NewAttributeSet(actionProps)),
		Resource:  models.NewEntity(resType, "urn:feature:1", models.NewAttributeSet(resProps)),
		Context:   models.NewAttributeSet(nil),
	}
}

const (
	gmwLayer = "https://api.example.gov.nl/bro/gmw/v1/collections/putten/items"
	bhrLayer = "https://api.example.gov.nl/bro/bhr/v1/wfs"
	cptLayer = "https://api.example.gov.nl/bro/cpt/v1/collections/sonderingen/items"
	rdCRS    = "http://www.opengis.net/def/crs/EPSG/0/28992"
)

func obligations(t *testing.T, resp *models.Response) []map[string]any {
	t.Helper()
	if resp.Attributes == nil {
		return nil
	}
	raw, ok := resp.Attributes["obligations"]
	if !ok {
		return nil
	}
	obs, ok := raw.([]map[string]any)
	if !ok {
		t.Fatalf("obligations wrong type: %T", raw)
	}
	return obs
}

// --- Example 1: feature-property sensitivity + waivable government access ----

func TestExample1_PublicLowSensitivity_Permit(t *testing.T) {
	c := newController(t, nil, "example1.ttl")
	resp, err := c.Authorize("t", req("urn:org:public", "", gmwLayer, rdCRS,
		"POINT (155000 463000)", map[string]any{"gevoeligheid": 2}))
	if err != nil || !resp.Allowed {
		t.Fatalf("want permit, got allowed=%v msg=%q err=%v", resp.Allowed, resp.Message, err)
	}
}

func TestExample1_PublicHighSensitivity_Deny(t *testing.T) {
	c := newController(t, nil, "example1.ttl")
	resp, _ := c.Authorize("t", req("urn:org:public", "", gmwLayer, rdCRS,
		"POINT (155000 463000)", map[string]any{"gevoeligheid": 5}))
	if resp.Allowed {
		t.Fatalf("want deny for sensitivity>=4 without government assignee, got permit")
	}
}

func TestExample1_GovernmentAssignee_WaivedPermit(t *testing.T) {
	c := newController(t, nil, "example1.ttl")
	// the waterschap has an Agreement (assignee oorg41000) that waives the
	// sensitivity constraint -> sees sensitive features.
	resp, _ := c.Authorize("t", req(
		"https://identifier.overheid.nl/tooi/id/oorg/oorg41000",
		"https://w3id.org/dpv#FulfilmentOfObligation",
		gmwLayer, rdCRS, "POINT (155000 463000)", map[string]any{"gevoeligheid": 5}))
	if !resp.Allowed {
		t.Fatalf("want permit for government assignee, got deny: %q", resp.Message)
	}
}

// --- Example 2: Defensie WKT polygon prohibition ----------------------------

func TestExample2_PointInside_Deny(t *testing.T) {
	c := newController(t, nil, "example2.ttl")
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (155000 463000)", nil)) // inside 150000..158000 x 460000..466000.
	if resp.Allowed {
		t.Fatalf("want deny for point inside Defensie area, got permit")
	}
}

func TestExample2_PointOutside_Permit(t *testing.T) {
	c := newController(t, nil, "example2.ttl")
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (100000 400000)", nil)) // far outside.
	if !resp.Allowed {
		t.Fatalf("want permit for point outside Defensie area, got deny: %q", resp.Message)
	}
}

func TestExample2_PointOnBoundary_Permit(t *testing.T) {
	c := newController(t, nil, "example2.ttl")
	// exactly on the western edge: sfWithin is false on the boundary (spec §3.2),
	// so the prohibition does not fire.
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (150000 463000)", nil))
	if !resp.Allowed {
		t.Fatalf("want permit for boundary point (not within), got deny: %q", resp.Message)
	}
}

// --- Example 3: "ja, mits" obligations --------------------------------------

func TestExample3_JaMits_PermitWithObligations(t *testing.T) {
	c := newController(t, nil, "example3.ttl")
	resp, _ := c.Authorize("t", req(
		"https://identifier.overheid.nl/tooi/id/oorg/oorg55500",
		"https://w3id.org/dpv#ResearchAndDevelopment",
		cptLayer, rdCRS, "POINT (155000 463000)", nil))
	if !resp.Allowed {
		t.Fatalf("want permit, got deny: %q", resp.Message)
	}
	obs := obligations(t, resp)
	if len(obs) != 2 {
		t.Fatalf("want 2 obligations, got %d: %+v", len(obs), obs)
	}
	ids := map[string]map[string]any{}
	for _, o := range obs {
		ids[o["id"].(string)] = o["properties"].(map[string]any)
	}
	gen, ok := ids[geonlNS+"generaliseren"]
	if !ok {
		t.Fatalf("missing generaliseren obligation: %+v", obs)
	}
	if gen["schaalnoemer"] != int64(50000) {
		t.Errorf("schaalnoemer = %v, want 50000", gen["schaalnoemer"])
	}
	nd, ok := ids[geonlNS+"nietDoorleveren"]
	if !ok {
		t.Fatalf("missing nietDoorleveren obligation: %+v", obs)
	}
	if nd["reikwijdte"] != "derden" {
		t.Errorf("reikwijdte = %v, want derden", nd["reikwijdte"])
	}
}

func TestExample3_WrongPurpose_Deny(t *testing.T) {
	c := newController(t, nil, "example3.ttl")
	// no processing_activity -> purpose refinement not satisfied -> default deny.
	resp, _ := c.Authorize("t", req(
		"https://identifier.overheid.nl/tooi/id/oorg/oorg55500", "",
		cptLayer, rdCRS, "POINT (155000 463000)", nil))
	if resp.Allowed {
		t.Fatal("want deny when purpose refinement unmet, got permit")
	}
}

// --- LayerSelection resolved via PIP ----------------------------------------

func zoneEntity(id, classificatie, wkt string) models.Entity {
	return models.NewEntity(
		"https://api.example.gov.nl/geo/milieuzonering", id,
		models.NewAttributeSet(map[string]any{"classificatie": classificatie, "geometry": wkt}),
	)
}

func TestLayerSelection_InsideSensitiveZone_Deny(t *testing.T) {
	zone := zoneEntity("z1", "gevoelig", square())
	c := newController(t, []models.Entity{zone}, "layerselection.ttl")
	// point inside the sensitive zone -> not disjoint -> permission inactive -> deny.
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (155000 463000)", nil))
	if resp.Allowed {
		t.Fatal("want deny for point inside sensitive zone, got permit")
	}
}

func TestLayerSelection_OutsideZone_Permit(t *testing.T) {
	zone := zoneEntity("z1", "gevoelig", square())
	c := newController(t, []models.Entity{zone}, "layerselection.ttl")
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (100000 400000)", nil)) // outside -> disjoint -> permit.
	if !resp.Allowed {
		t.Fatalf("want permit for point outside sensitive zone, got deny: %q", resp.Message)
	}
}

func TestLayerSelection_NonSensitiveZoneFilteredOut_Permit(t *testing.T) {
	// the only zone is NOT 'gevoelig' -> filtered out by where -> empty selection
	// -> disjoint vacuously true -> permit even for an inside point.
	zone := zoneEntity("z1", "regulier", square())
	c := newController(t, []models.Entity{zone}, "layerselection.ttl")
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (155000 463000)", nil))
	if !resp.Allowed {
		t.Fatalf("want permit (no sensitive zones), got deny: %q", resp.Message)
	}
}

func TestLayerSelection_Unresolvable_FailSafeDeny(t *testing.T) {
	// no zone entities at all -> layer not resolvable -> not-evaluable ->
	// permission inactive -> fail-safe default deny (spec §7.5).
	c := newController(t, nil, "layerselection.ttl")
	resp, _ := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (155000 463000)", nil))
	if resp.Allowed {
		t.Fatal("want fail-safe deny when layer unresolvable, got permit")
	}
}

// square returns the Defensie/zone example polygon as a CRS-prefixed WKT literal.
func square() string {
	return "<http://www.opengis.net/def/crs/EPSG/0/28992> POLYGON ((150000 460000, 158000 460000, 158000 466000, 150000 466000, 150000 460000))"
}

// --- Tile evaluation --------------------------------------------------------

// filterFeaturesProps returns the properties of the first geonl:filterFeatures
// obligation in the response.
func filterFeaturesProps(t *testing.T, resp *models.Response) map[string]any {
	t.Helper()
	for _, o := range obligations(t, resp) {
		if o["id"] == geonlFilterFeaturesAction {
			props, _ := o["properties"].(map[string]any)
			return props
		}
	}
	return nil
}

func TestTile_Straddling_PermitWithFilterFeatures(t *testing.T) {
	c := newController(t, nil, "example2.ttl")
	// tile envelope straddling the Defensie polygon boundary (partly in, partly out).
	env := "POLYGON ((155000 463000, 165000 463000, 165000 470000, 155000 470000, 155000 463000))"
	resp, _ := c.Authorize("t", typedReq(resourceTile, "urn:org:x", "", bhrLayer, rdCRS, env, nil))
	if !resp.Allowed {
		t.Fatalf("want permit for mixed tile, got deny: %q", resp.Message)
	}
	props := filterFeaturesProps(t, resp)
	if props == nil {
		t.Fatalf("want geonl:filterFeatures obligation on mixed tile, got %+v", obligations(t, resp))
	}
	// A4/B1: the obligation must carry the spatial predicate the BRO PEP reads,
	// not empty properties.
	if props["spatial"] != "notWithin" {
		t.Errorf("spatial = %v, want notWithin", props["spatial"])
	}
	wkt, _ := props["wkt"].(string)
	if !strings.HasPrefix(wkt, "POLYGON") || !strings.Contains(wkt, "150000 460000") {
		t.Errorf("wkt = %q, want the Defensie polygon in RD", wkt)
	}
	if props["crs"] == nil || props["crs"] == "" {
		t.Errorf("crs missing from spatial filterFeatures predicate")
	}
}

// TestTile_GMWFeatureProperty_PermitWithFilterFeatures covers B2: a
// geonl:featureProperty constraint that cannot be tested per-tile must not
// fail-closed to deny but permit with a filterFeatures obligation carrying the
// property predicate the PEP can turn into a SQL/CQL filter.
func TestTile_GMWFeatureProperty_PermitWithFilterFeatures(t *testing.T) {
	c := newController(t, nil, "example1.ttl")
	// tile request for a non-government subject: the Offer permission is gated by
	// "gevoeligheid < 4", which has no per-feature attribute on a tile.
	resp, _ := c.Authorize("t", typedReq(resourceTile, "urn:org:public", "", gmwLayer, rdCRS,
		"POLYGON ((150000 460000, 160000 460000, 160000 466000, 150000 466000, 150000 460000))", nil))
	if !resp.Allowed {
		t.Fatalf("want permit for GMW tile (filter per feature), got deny: %q", resp.Message)
	}
	props := filterFeaturesProps(t, resp)
	if props == nil {
		t.Fatalf("want geonl:filterFeatures obligation on GMW tile, got %+v", obligations(t, resp))
	}
	if props["property"] != "gevoeligheid" {
		t.Errorf("property = %v, want gevoeligheid", props["property"])
	}
	if props["operator"] != "lt" {
		t.Errorf("operator = %v, want lt", props["operator"])
	}
	if props["value"] != int64(4) {
		t.Errorf("value = %#v, want int64(4)", props["value"])
	}
	// B4: when a tile is permitted with a filter-obligation because no
	// whole-tile permission was active, the audit must credit the real policy
	// that carried the constraint, never a synthetic "tile" placeholder.
	const wantPolicy = "https://example.gov.nl/bro/gmw/aanbod"
	if resp.PolicyKey != wantPolicy {
		t.Errorf("PolicyKey = %q, want the real offer uid %q (not \"tile\")", resp.PolicyKey, wantPolicy)
	}
}

func TestTile_FullyInside_Deny(t *testing.T) {
	c := newController(t, nil, "example2.ttl")
	env := "POLYGON ((151000 461000, 157000 461000, 157000 465000, 151000 465000, 151000 461000))"
	resp, _ := c.Authorize("t", typedReq(resourceTile, "urn:org:x", "", bhrLayer, rdCRS, env, nil))
	if resp.Allowed {
		t.Fatal("want deny for tile fully inside prohibited area, got permit")
	}
}

// TestMerge_DedupPoliciesAcrossEnvelopes covers A10: the importer stores the
// full source in every per-policy envelope, so the geo loader re-parses all
// policies from each envelope. The engine must deduplicate policies by uid
// across envelopes, otherwise a request would collect duplicate obligations.
func TestMerge_DedupPoliciesAcrossEnvelopes(t *testing.T) {
	ctx := context.Background()
	logger := testLogger()
	p := pap.New(ctx, logger)

	src, err := os.ReadFile("testdata/example3.ttl")
	if err != nil {
		t.Fatal(err)
	}
	// store the SAME document under two PAP ids, as the importer does when a
	// document holds several policies (one envelope per policy, full source each).
	for _, id := range []string{"env-a", "env-b"} {
		stored, err := pap.NewPolicyFromData(id, "odrl", "", "urn:"+id, bytes.NewReader(wrapEnvelope(t, src)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.Create(stored); err != nil {
			t.Fatal(err)
		}
	}
	ip := pip.New(ctx, logger)
	c := NewController(pdp.WithContext(ctx), pdp.WithLogger(logger), pdp.WithPAP(p), pdp.WithPIP(ip))

	resp, _ := c.Authorize("t", req(
		"https://identifier.overheid.nl/tooi/id/oorg/oorg55500",
		"https://w3id.org/dpv#ResearchAndDevelopment",
		cptLayer, rdCRS, "POINT (155000 463000)", nil))
	if !resp.Allowed {
		t.Fatalf("want permit, got deny: %q", resp.Message)
	}
	obs := obligations(t, resp)
	if len(obs) != 2 {
		t.Fatalf("want 2 obligations (deduplicated across 2 envelopes), got %d: %+v", len(obs), obs)
	}
}

// TestConsultedSources_LayerSelection proves ingreep 3: a LayerSelection evaluation
// records the exact PIP sources it consulted (the selected layer and each zone entity
// with the attributes read off it) in Response.Attributes["consultedSources"].
func TestConsultedSources_LayerSelection(t *testing.T) {
	zone := zoneEntity("z1", "gevoelig", square())
	c := newController(t, []models.Entity{zone}, "layerselection.ttl")
	resp, err := c.Authorize("t", req("urn:org:x", "", bhrLayer, rdCRS,
		"POINT (155000 463000)", nil))
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := resp.Attributes["consultedSources"]
	if !ok {
		t.Fatal("expected consultedSources in response attributes")
	}
	keys, ok := raw.([]string)
	if !ok {
		t.Fatalf("consultedSources wrong type: %T", raw)
	}
	set := map[string]bool{}
	for _, k := range keys {
		set[k] = true
	}
	if !set["https://api.example.gov.nl/geo/milieuzonering"] {
		t.Errorf("expected the selected layer to be recorded as consulted; got %v", keys)
	}
	if !set["https://api.example.gov.nl/geo/milieuzonering::z1"] {
		t.Errorf("expected the consulted zone entity uid; got %v", keys)
	}
	if !set["https://api.example.gov.nl/geo/milieuzonering::z1/geometry"] {
		t.Errorf("expected the consulted geometry attribute; got %v", keys)
	}
}
