package models

import "strings"

// FTV AuthZEN informatiemodel: canonical Linked Data URIs for actions, subject-
// and resource-types, following the NLGov profile of the OpenID AuthZEN
// Authorization API (subject/resource `type` SHOULD be a Linked Data URI).
//
// The URIs below are the canonical form used for policy matching. Existing short
// names (e.g. "raadplegen", "user", "service") keep working: the Normalize* helpers
// translate a known short name (or synonym) to its canonical URI and leave any
// unknown value (or value that is already a URI) unchanged (backward compatible).

// Namespaces for the FTV informatiemodel ontology.
const (
	NSFTVDef      = "https://standaarden.overheid.nl/ftv/def/"
	NSFTVActie    = NSFTVDef + "actie/"    // AVG verwerkingshandelingen (actions).
	NSFTVSubject  = NSFTVDef + "subject/"  // subject (principal) types.
	NSFTVResource = NSFTVDef + "resource/" // resource (object) types.
)

// Canonical action URIs, modelled as the AVG (GDPR art. 4 sub 2) verwerkingshandelingen.
//
// ActieVerwerken is the root class; the others are its sub-actions. ActieVerstrekken
// groups the "verstrekken door middel van doorzending"-family (doorzenden, verspreiden,
// beschikbaar stellen) and ActieSamenbrengen groups met-elkaar-in-verband-brengen.
const (
	ActieVerwerken          = NSFTVActie + "verwerken" // root: any processing.
	ActieVerzamelen         = NSFTVActie + "verzamelen"
	ActieVastleggen         = NSFTVActie + "vastleggen"
	ActieOrdenen            = NSFTVActie + "ordenen"
	ActieStructureren       = NSFTVActie + "structureren"
	ActieOpslaan            = NSFTVActie + "opslaan"
	ActieBijwerken          = NSFTVActie + "bijwerken-of-wijzigen"
	ActieOpvragen           = NSFTVActie + "opvragen"
	ActieRaadplegen         = NSFTVActie + "raadplegen"
	ActieGebruiken          = NSFTVActie + "gebruiken"
	ActieVerstrekken        = NSFTVActie + "verstrekken" // grouping of doorzenden/verspreiden/beschikbaar-stellen.
	ActieDoorzenden         = NSFTVActie + "doorzenden"
	ActieVerspreiden        = NSFTVActie + "verspreiden"
	ActieBeschikbaarStellen = NSFTVActie + "beschikbaar-stellen"
	ActieSamenbrengen       = NSFTVActie + "samenbrengen"
	ActieInVerbandBrengen   = NSFTVActie + "met-elkaar-in-verband-brengen"
	ActieAfschermen         = NSFTVActie + "afschermen"
	ActieWissen             = NSFTVActie + "wissen"
	ActieVernietigen        = NSFTVActie + "vernietigen"
)

// Canonical subject (principal) type URIs. These align with the principal types in
// eam/pep/principal.go plus the NLGov transport-specific types (fsc-peer, ip-address).
const (
	SubjectUser        = NSFTVSubject + "user"
	SubjectApp         = NSFTVSubject + "app"
	SubjectDoelbinding = NSFTVSubject + "doelbinding"
	SubjectActivity    = NSFTVSubject + "activity" // RVVA / verwerkingsactiviteit.
	SubjectZaak        = NSFTVSubject + "zaak"
	SubjectFSCPeer     = NSFTVSubject + "fsc-peer"
	SubjectIPAddress   = NSFTVSubject + "ip-address"
)

// Canonical resource (object) type URIs.
const (
	ResourceResource = NSFTVResource + "resource"
	ResourceService  = NSFTVResource + "service"
	ResourceURI      = NSFTVResource + "uri"
	ResourceZaak     = NSFTVResource + "zaak"
)

// actionURIByAlias maps a lower-cased short name or synonym to its canonical action URI.
// Keys include the AVG verwerkingshandelingen, OpenFTV/informatiemodel synonyms
// (registreren, verstrekken, verwijderen, doorleveren, bijwerken) and HTTP methods.
var actionURIByAlias = map[string]string{
	// AVG verwerkingshandelingen.
	"verzamelen":                    ActieVerzamelen,
	"vastleggen":                    ActieVastleggen,
	"ordenen":                       ActieOrdenen,
	"structureren":                  ActieStructureren,
	"opslaan":                       ActieOpslaan,
	"bijwerken of wijzigen":         ActieBijwerken,
	"opvragen":                      ActieOpvragen,
	"raadplegen":                    ActieRaadplegen,
	"gebruiken":                     ActieGebruiken,
	"doorzenden":                    ActieDoorzenden,
	"verspreiden":                   ActieVerspreiden,
	"beschikbaar stellen":           ActieBeschikbaarStellen,
	"samenbrengen":                  ActieSamenbrengen,
	"met elkaar in verband brengen": ActieInVerbandBrengen,
	"afschermen":                    ActieAfschermen,
	"wissen":                        ActieWissen,
	"vernietigen":                   ActieVernietigen,
	"verwerken":                     ActieVerwerken,
	"verstrekken":                   ActieVerstrekken,
	// Synonyms / OpenFTV short names.
	"registreren": ActieVastleggen,
	"bijwerken":   ActieBijwerken,
	"wijzigen":    ActieBijwerken,
	"verwijderen": ActieWissen,
	"doorleveren": ActieDoorzenden,
	"lezen":       ActieRaadplegen,
	// HTTP methods (fallback action names, see informatiemodel).
	"get":     ActieRaadplegen,
	"head":    ActieRaadplegen,
	"options": ActieRaadplegen,
	"post":    ActieVastleggen,
	"put":     ActieBijwerken,
	"patch":   ActieBijwerken,
	"delete":  ActieWissen,
}

// subjectURIByAlias maps a lower-cased short subject/principal type to its canonical URI.
var subjectURIByAlias = map[string]string{
	"user":        SubjectUser,
	"app":         SubjectApp,
	"application": SubjectApp,
	"doelbinding": SubjectDoelbinding,
	"activity":    SubjectActivity,
	"rvva":        SubjectActivity,
	"zaak":        SubjectZaak,
	"fsc-peer":    SubjectFSCPeer,
	"fsc_peer":    SubjectFSCPeer,
	"ip-address":  SubjectIPAddress,
	"ip_address":  SubjectIPAddress,
}

// resourceURIByAlias maps a lower-cased short resource/object type to its canonical URI.
var resourceURIByAlias = map[string]string{
	"resource": ResourceResource,
	"service":  ResourceService,
	"uri":      ResourceURI,
	"zaak":     ResourceZaak,
}

// NormalizeActionURI returns the canonical action URI for the given action name.
//
// A value that is already an FTV action URI, or that is unknown, is returned unchanged.
func NormalizeActionURI(name string) string {
	if name == "" || strings.HasPrefix(name, NSFTVActie) {
		return name
	}
	if uri, ok := actionURIByAlias[strings.ToLower(strings.TrimSpace(name))]; ok {
		return uri
	}
	return name
}

// NormalizeSubjectTypeURI returns the canonical subject (principal) type URI.
//
// A value that is already an FTV subject URI, or that is unknown, is returned unchanged.
func NormalizeSubjectTypeURI(t string) string {
	if t == "" || strings.HasPrefix(t, NSFTVSubject) {
		return t
	}
	if uri, ok := subjectURIByAlias[strings.ToLower(strings.TrimSpace(t))]; ok {
		return uri
	}
	return t
}

// NormalizeResourceTypeURI returns the canonical resource (object) type URI.
//
// A value that is already an FTV resource URI, or that is unknown, is returned unchanged.
func NormalizeResourceTypeURI(t string) string {
	if t == "" || strings.HasPrefix(t, NSFTVResource) {
		return t
	}
	if uri, ok := resourceURIByAlias[strings.ToLower(strings.TrimSpace(t))]; ok {
		return uri
	}
	return t
}

// NormalizeEntityTypeURI returns the canonical URI for a generic entity type, trying
// the subject-type table first and then the resource-type table. Because "zaak" is
// both a subject and a resource type it resolves to the subject URI here; use
// NormalizeSubjectTypeURI / NormalizeResourceTypeURI when the role is known.
//
// A value that is already an FTV subject/resource URI, or that is unknown, is returned unchanged.
func NormalizeEntityTypeURI(t string) string {
	if uri := NormalizeSubjectTypeURI(t); uri != t {
		return uri
	}
	return NormalizeResourceTypeURI(t)
}
