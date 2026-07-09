// Package export describes the policies administered by a PAP as an
// ODRL-AP-NL document: every executable policy (Rego/Cedar/OpenFGA) becomes an
// apnl:PolicyArtifact, bundled under an odrl:Offer/Set whose permissions carry
// apnl:conformsToPolicy constraints referring to those artifacts.
package export

import (
	"errors"
	"io/fs"
	"os"

	"github.com/goccy/go-yaml"
)

// AnnotationsFile is the default name of the export annotation file,
// located next to (inside) the policy store directory.
const AnnotationsFile = "odrl-export.yaml"

// Annotations is the (optional) odrl-export.yaml document: administrative
// metadata that cannot be derived from the policy files themselves.
//
// Example:
//
//	uidBase: "https://pap.example.gov.nl/odrl/"
//	policy:
//	  type: Offer
//	  title: "Aanbod toeslagen-gegevens"
//	  publisher: "https://identifier.overheid.nl/tooi/id/oorg/oorg10103"
//	  assigner: "https://identifier.overheid.nl/tooi/id/oorg/oorg10103"
//	  issued: "2026-07-08"
//	  target: "https://data.example.gov.nl/toeslagen/distributie"
//	policies:
//	  opa/dienst_toeslagen/zorgtoeslag:
//	    title: "Doelbindingsregels zorgtoeslag"
//	    purpose: "https://register.example.gov.nl/verwerkingen/zorgtoeslag"
//	    target: "https://data.example.gov.nl/toeslagen/distributie"
//	    velden:
//	      - "https://data.rijksoverheid.nl/brp/rubriek/geslachtsnaamPersoon"
//	    voorwaarden:
//	      - leftOperand: "datumOverlijdenOverlijden"   # rubriek-naam of volledige IRI
//	        operator: "knv"                            # kort (knv/ga1) of def-IRI
//	      - leftOperand: "gemeenteVanInschrijvingVerblijfplaats"
//	        operator: "ga1"
//	        rightOperand: "https://identifier.overheid.nl/tooi/id/gemeente/gm0014"
type Annotations struct {
	// UIDBase is the IRI prefix for generated policy/artifact uids.
	UIDBase string `yaml:"uidBase"`
	// Policy carries set-level metadata for the exported Offer/Set.
	Policy PolicyMeta `yaml:"policy"`
	// Policies maps a PAP policy key (e.g. "opa/dienst_toeslagen/zorgtoeslag")
	// to per-artifact metadata.
	Policies map[string]ArtifactMeta `yaml:"policies"`
}

// PolicyMeta is the set-level metadata of the exported Offer/Set.
type PolicyMeta struct {
	UID          string `yaml:"uid"`          // optional explicit uid.
	Type         string `yaml:"type"`         // "Offer" (default), "Set" or "Agreement".
	Title        string `yaml:"title"`        // dct:title (Dutch).
	Description  string `yaml:"description"`  // dct:description (Dutch).
	Publisher    string `yaml:"publisher"`    // dct:publisher (TOOI IRI).
	Assigner     string `yaml:"assigner"`     // odrl:assigner.
	Assignee     string `yaml:"assignee"`     // odrl:assignee (Agreements).
	Issued       string `yaml:"issued"`       // dct:issued (xsd:date).
	Target       string `yaml:"target"`       // default odrl:target (dataset/distribution).
	Instantiates string `yaml:"instantiates"` // apnl:instantiates (Agreements).
}

// ArtifactMeta is per-policy-key metadata for the exported artifact and its
// permission.
type ArtifactMeta struct {
	Title       string `yaml:"title"`       // dct:title of the artifact.
	Description string `yaml:"description"` // dct:description of the artifact.
	Target      string `yaml:"target"`      // odrl:target for this permission.
	Purpose     string `yaml:"purpose"`     // purpose IRI: refinement rightOperand.
	Action      string `yaml:"action"`      // odrl action IRI (default odrl:use).
	Entrypoint  string `yaml:"entrypoint"`  // apnl:entrypoint override.
	DownloadURL string `yaml:"downloadURL"` // dcat:downloadURL override.
	// Velden are the requested BRP field-group (rubriek) IRIs for this policy.
	// Each is emitted both as brp:verzochteRubriek and as odrl:target on the
	// permission, so the RvIG harvester can derive the requested/authorised
	// fields from the export instead of getting an empty union (finding C1).
	Velden []string `yaml:"velden"`
	// Voorwaarden are the record-bound condition rules (voorwaarderegels) for
	// this policy. Each is emitted as an odrl:constraint on the permission — in
	// the exact blank-node shape the BRP-sim harvester recognises (leftOperand
	// rubriek-IRI, operator brp-def-IRI, optional rightOperand) — so an
	// export-harvested profile keeps the WOZ-style rules (e.g. brp:knv "geen
	// overleden personen") instead of silently dropping them.
	Voorwaarden []VoorwaardeMeta `yaml:"voorwaarden"`
}

// VoorwaardeMeta is one record-bound condition rule (odrl:constraint) on a
// policy's permission.
//
// LeftOperand and Operator accept either a full IRI or a short name: a bare
// LeftOperand is resolved against the BRP rubriek namespace (brprub:), a bare
// Operator (e.g. "knv", "ga1") against the BRP def namespace (brp:). The
// optional RightOperand is emitted as an IRI reference when it looks like an
// IRI (contains "://") and as a plain literal otherwise.
type VoorwaardeMeta struct {
	LeftOperand  string `yaml:"leftOperand"`  // rubriek-naam or full IRI.
	Operator     string `yaml:"operator"`     // short name (knv/ga1) or full IRI.
	RightOperand string `yaml:"rightOperand"` // optional IRI or literal.
}

// LoadAnnotations reads the annotation file from the given path.
// A missing file is not an error: the annotations are optional and an empty
// set is returned in that case.
func LoadAnnotations(path string) (*Annotations, error) {
	a := &Annotations{}
	if path == "" {
		return a, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return a, nil
		}
		return nil, err
	}

	if err = yaml.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}
