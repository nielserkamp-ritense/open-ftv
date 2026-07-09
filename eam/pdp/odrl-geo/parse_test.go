package odrl_geo

import (
	"bytes"
	"encoding/base64"
	"os"
	"testing"

	"github.com/goccy/go-json"
)

// wrapEnvelope wraps raw turtle bytes in a minimal PAP envelope (matching the
// importer's storage format) so loadDocument can be exercised directly.
func wrapEnvelope(t *testing.T, ttl []byte) []byte {
	t.Helper()
	env := map[string]any{
		"uid":      "test",
		"mimeType": "text/turtle",
		"source":   base64.StdEncoding.EncodeToString(ttl),
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func loadTestdata(t *testing.T, name string) *document {
	t.Helper()
	ttl, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := loadDocument(wrapEnvelope(t, ttl))
	if err != nil {
		t.Fatalf("loadDocument(%s): %v", name, err)
	}
	return doc
}

func TestParseDefensiePolicy(t *testing.T) {
	doc := loadTestdata(t, "example2.ttl")

	// the named Defensie area must be resolved with its WKT geometry.
	var found *area
	for _, a := range doc.Areas {
		if bytes.Contains([]byte(a.UID), []byte("defensiegebiedX")) {
			found = a
		}
	}
	if found == nil {
		t.Fatal("defensiegebiedX area not parsed")
	}
	if !bytes.Contains([]byte(found.WKT), []byte("POLYGON")) {
		t.Errorf("area WKT missing polygon: %q", found.WKT)
	}

	// the prohibition must carry a spatial constraint referencing the area.
	var prohib *geoConstraint
	for _, p := range doc.Policies {
		for _, r := range p.Prohibitions {
			for _, cons := range r.Constraints {
				if cons.isSpatial() {
					prohib = cons
				}
			}
		}
	}
	if prohib == nil {
		t.Fatal("no spatial prohibition constraint parsed")
	}
	if prohib.Operator != sfWithin {
		t.Errorf("operator = %q, want sfWithin", prohib.Operator)
	}
	if prohib.RightRef == "" {
		t.Error("rightOperandReference not captured (geo predicate dropped)")
	}
}

func TestParseFeaturePropertyConstraint(t *testing.T) {
	doc := loadTestdata(t, "example1.ttl")

	var c *geoConstraint
	for _, p := range doc.Policies {
		for _, r := range p.Permissions {
			for _, cons := range r.Constraints {
				if cons.LeftOperand == geonlFeatureProperty {
					c = cons
				}
			}
		}
	}
	if c == nil {
		t.Fatal("featureProperty constraint not parsed")
	}
	if attributeKey(c.Property) != "gevoeligheid" {
		t.Errorf("property key = %q, want gevoeligheid", attributeKey(c.Property))
	}
	if c.Operator != opLt || len(c.RightLiterals) != 1 || c.RightLiterals[0].Value != "4" {
		t.Errorf("constraint = %+v, want lt 4", c)
	}
	if !c.Waivable {
		t.Error("waivable flag not captured")
	}
}

func TestParseDuties(t *testing.T) {
	doc := loadTestdata(t, "example3.ttl")

	var duties []*geoDuty
	for _, p := range doc.Policies {
		for _, r := range p.Permissions {
			duties = append(duties, r.Duties...)
		}
	}
	if len(duties) == 0 {
		t.Fatal("no duties parsed")
	}

	var gen *geoDuty
	for _, d := range duties {
		if d.Action == geonlNS+"generaliseren" {
			gen = d
		}
	}
	if gen == nil {
		t.Fatal("generaliseren duty not parsed")
	}
	ob := dutyToObligation(gen)
	props := ob["properties"].(map[string]any)
	if props["schaalnoemer"] != int64(50000) {
		t.Errorf("schaalnoemer = %v (%T), want 50000", props["schaalnoemer"], props["schaalnoemer"])
	}
}

func TestParseLayerSelection(t *testing.T) {
	doc := loadTestdata(t, "layerselection.ttl")
	if len(doc.Selections) != 1 {
		t.Fatalf("selections = %d, want 1", len(doc.Selections))
	}
	for _, sel := range doc.Selections {
		if sel.FromLayer != "https://api.example.gov.nl/geo/milieuzonering" {
			t.Errorf("fromLayer = %q", sel.FromLayer)
		}
		if sel.Where == nil || sel.Where.Operator != opEq {
			t.Errorf("where filter not parsed: %+v", sel.Where)
		}
	}
}
