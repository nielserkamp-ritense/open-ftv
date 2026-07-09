package doelbinding

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"

// Configuration names under which this plugin registers its mappers.
// They match the names accepted by the historic hardcoded switch, so existing
// REQUEST_MAPPINGS configuration keeps working unchanged.
const (
	NameDoelbindingToPrincipal = "doelbindingtoprincipal"
	NameRvvaToPrincipal        = "rvvatoprincipal"
)

// init registers the doelbinding-family mappers with the core registry, so that
// a blank/underscore import of this package makes them resolvable by name.
func init() {
	mapping.Register(NameDoelbindingToPrincipal, DoelbindingToPrincipal)
	mapping.Register(NameRvvaToPrincipal, RvvaToPrincipal)
}
