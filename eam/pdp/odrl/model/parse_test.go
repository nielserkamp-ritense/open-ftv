package model

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const examples = "../../../../testdata/odrl/examples"

func load(t *testing.T, name, mime string) *Document {
	t.Helper()

	f, err := os.Open(filepath.Join(examples, name))
	require.NoError(t, err)
	defer f.Close()

	d, err2 := Parse(f, mime)
	require.NoError(t, err2)
	return d
}

func TestParseExample1Turtle(t *testing.T) {
	t.Parallel()

	d := load(t, "1-generiek-drietraps.ttl", "text/turtle")

	require.Len(t, d.Policies, 3) // 1 Offer + 2 Agreements.

	offer := d.Policy("https://example.gov.nl/brf/brfAanbod")
	require.NotNil(t, offer)
	assert.Equal(t, "http://www.w3.org/ns/odrl/2/Offer", offer.Type)
	assert.Equal(t, APNLProfile, offer.Profile)
	assert.Equal(t, "https://identifier.overheid.nl/tooi/id/oorg/oorg99001", offer.Publisher)
	assert.Equal(t, "2026-01-15", offer.Issued)
	require.Len(t, offer.Permissions, 1)
	assert.Len(t, offer.Permissions[0].Constraints, 2)
	require.Len(t, offer.Obligations, 1)
	assert.Equal(t, "https://example.gov.nl/brf/duty-protocollering", offer.Obligations[0].Ref)
	require.NotEmpty(t, offer.Titles)
	assert.Equal(t, "Aanbod BRF-gegevens", offer.Titles[0].Value)
	assert.Equal(t, "nl", offer.Titles[0].Language)

	utrecht := d.Policy("https://example.gov.nl/brf/overeenkomst-utrecht")
	require.NotNil(t, utrecht)
	assert.Equal(t, "http://www.w3.org/ns/odrl/2/Agreement", utrecht.Type)
	assert.Equal(t, "https://example.gov.nl/brf/brfAanbod", utrecht.Instantiates)
	assert.Equal(t, "https://example.gov.nl/brf/brfAanbod", utrecht.WasDerivedFrom)
	assert.Equal(t, "https://identifier.overheid.nl/tooi/id/gemeente/gm0344", utrecht.Assignee)
	require.Len(t, utrecht.Permissions, 2)

	// purpose refinement on the action.
	var found bool
	for _, perm := range utrecht.Permissions {
		require.NotNil(t, perm.Action)
		assert.Equal(t, "http://www.w3.org/ns/odrl/2/read", perm.Action.Value)
		for _, ref := range perm.Action.Refinements {
			if len(ref.RightOperand) == 1 &&
				ref.RightOperand[0].ID == "https://register.utrecht.nl/verwerkingen/handhaving-openbare-ruimte" {
				assert.Equal(t, "http://www.w3.org/ns/odrl/2/purpose", ref.LeftOperand)
				assert.Equal(t, "http://www.w3.org/ns/odrl/2/isA", ref.Operator)
				found = true
			}
		}
	}
	assert.True(t, found, "purpose refinement should be parsed")

	// named constraints with waivable markers.
	require.Len(t, d.Constraints, 2)
	byUID := map[string]*Constraint{}
	for _, c := range d.Constraints {
		byUID[c.UID] = c
	}

	eigen := byUID["https://example.gov.nl/brf/alleenEigenGemeente"]
	require.NotNil(t, eigen)
	require.NotNil(t, eigen.Waivable)
	assert.True(t, *eigen.Waivable)

	open := byUID["https://example.gov.nl/brf/alleenOpenstaandeAangiften"]
	require.NotNil(t, open)
	require.NotNil(t, open.Waivable)
	assert.False(t, *open.Waivable)
	require.Len(t, open.RightOperand, 1)
	assert.Equal(t, "open", open.RightOperand[0].Value)

	// scope waiving: the politie agreement omits the waivable constraint.
	politie := d.Policy("https://example.gov.nl/brf/overeenkomst-politie")
	require.NotNil(t, politie)
	for _, perm := range politie.Permissions {
		require.Len(t, perm.Constraints, 1)
		assert.Equal(t, "https://example.gov.nl/brf/alleenOpenstaandeAangiften", perm.Constraints[0].Ref)
	}
}

func TestParseExample1JSONLD(t *testing.T) {
	t.Parallel()

	d := load(t, "1-generiek-drietraps.jsonld", "application/ld+json")

	require.Len(t, d.Policies, 2) // JSON-LD variant has Offer + 1 Agreement.
	offer := d.Policy("https://example.gov.nl/brf/brfAanbod")
	require.NotNil(t, offer)
	assert.Equal(t, APNLProfile, offer.Profile)
	require.Len(t, offer.Permissions, 1)
	assert.Len(t, offer.Permissions[0].Constraints, 2)

	utrecht := d.Policy("https://example.gov.nl/brf/overeenkomst-utrecht")
	require.NotNil(t, utrecht)
	require.Len(t, utrecht.Permissions, 2)
	assert.Equal(t, "https://example.gov.nl/brf/brfAanbod", utrecht.Instantiates)
}

func TestParseExample3ConformsToPolicy(t *testing.T) {
	t.Parallel()

	d := load(t, "3-conforms-to-policy.ttl", "text/turtle")

	require.Len(t, d.Artifacts, 1)
	a := d.Artifacts[0]
	assert.Equal(t, "https://example.gov.nl/brp/artefact-burgerzaken", a.UID)
	assert.Equal(t, APNLRegoModule, a.Type)
	assert.Equal(t, "application/vnd.rego", a.Format)
	assert.Equal(t, "data.doelbinding.burgerzaken.allow", a.Entrypoint)
	assert.Equal(t, "e1482e92c4cf0faa3c88bfdc54bc39912fba6d5f6e8db298bc364ac7d5e8bfa8", a.SHA256)
	assert.NotEmpty(t, a.DownloadURL)

	require.Len(t, d.Policies, 1)
	p := d.Policies[0]
	require.Len(t, p.Permissions, 1)
	perm := p.Permissions[0]
	assert.Len(t, perm.Targets, 5)

	require.Len(t, perm.Constraints, 1)
	c := perm.Constraints[0]
	assert.Equal(t, APNLVerwerkingsverzoek, c.LeftOperand)
	assert.Equal(t, APNLConformsToPolicy, c.Operator)
	require.Len(t, c.RightOperand, 1)
	assert.Equal(t, "https://example.gov.nl/brp/artefact-burgerzaken", c.RightOperand[0].ID)

	// domain extension (brp:medium) must be preserved in the Extra bag.
	var brpMedium bool
	for _, e := range perm.Extra {
		if e.Predicate == "https://data.rijksoverheid.nl/brp/def#medium" {
			brpMedium = true
		}
	}
	assert.True(t, brpMedium, "brp:medium should be kept as extra property")
}

func TestParseExample4Bundle(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, mime string }{
		{"4-policy-bundle.ttl", "text/turtle"},
		{"4-policy-bundle.jsonld", "application/ld+json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := load(t, tc.name, tc.mime)

			require.GreaterOrEqual(t, len(d.Bundles), 1)
			var v2 *Bundle
			for _, b := range d.Bundles {
				if b.UID == "https://example.gov.nl/brp/bundel-brp-doelbinding-v2" {
					v2 = b
				}
			}
			require.NotNil(t, v2)
			assert.Equal(t, "application/vnd.openpolicyagent.bundle", v2.Format)
			assert.Len(t, v2.Bundles, 2)
			assert.Equal(t, "https://example.gov.nl/brp/bundel-brp-doelbinding-v1", v2.WasRevisionOf)
			assert.Len(t, d.Artifacts, 2)
		})
	}
}

// TestRoundTrip parses each example, serializes it to Turtle and JSON-LD, and
// re-parses the result; the second parse must yield the same model.
func TestRoundTrip(t *testing.T) {
	t.Parallel()

	files := []struct{ name, mime string }{
		{"1-generiek-drietraps.ttl", "text/turtle"},
		{"1-generiek-drietraps.jsonld", "application/ld+json"},
		{"2-brp-drietraps.ttl", "text/turtle"},
		{"3-conforms-to-policy.ttl", "text/turtle"},
		{"4-policy-bundle.ttl", "text/turtle"},
		{"4-policy-bundle.jsonld", "application/ld+json"},
	}

	for _, tc := range files {
		for _, out := range []string{"text/turtle", "application/ld+json"} {
			t.Run(tc.name+"->"+out, func(t *testing.T) {
				t.Parallel()

				d1 := load(t, tc.name, tc.mime)

				var buf bytes.Buffer
				require.NoError(t, d1.Serialize(&buf, out))

				d2, err := Parse(&buf, out)
				require.NoError(t, err)

				assert.Equal(t, count(d1), count(d2), "entity counts must survive a round-trip")

				for _, p1 := range d1.Policies {
					p2 := d2.Policy(p1.UID)
					require.NotNil(t, p2, "policy %s must survive", p1.UID)
					assert.Equal(t, p1.Type, p2.Type)
					assert.Equal(t, p1.Publisher, p2.Publisher)
					assert.Equal(t, p1.Issued, p2.Issued)
					assert.Equal(t, p1.Instantiates, p2.Instantiates)
					assert.Equal(t, len(p1.Permissions), len(p2.Permissions))
					assert.Equal(t, len(p1.Obligations), len(p2.Obligations))
					assert.ElementsMatch(t, p1.Titles, p2.Titles)
				}

				for _, a1 := range d1.Artifacts {
					a2 := d2.Artifact(a1.UID)
					require.NotNil(t, a2, "artifact %s must survive", a1.UID)
					assert.Equal(t, a1.Format, a2.Format)
					assert.Equal(t, a1.SHA256, a2.SHA256)
					assert.Equal(t, a1.Entrypoint, a2.Entrypoint)
					assert.Equal(t, a1.DownloadURL, a2.DownloadURL)
				}
			})
		}
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	d := load(t, "1-generiek-drietraps.ttl", "text/turtle")
	assert.NoError(t, d.Validate())

	bad := &Document{Policies: []*Policy{{UID: "https://example.gov.nl/x"}}}
	err := bad.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dct:title")
	assert.Contains(t, err.Error(), "dct:publisher")
	assert.Contains(t, err.Error(), "odrl:profile")

	empty := &Document{}
	assert.Error(t, empty.Validate())
}

func count(d *Document) [5]int {
	return [5]int{len(d.Policies), len(d.Artifacts), len(d.Bundles), len(d.Constraints), len(d.Duties)}
}
