package body

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"

// NameBodyToContext is the configuration name (`REQUEST_MAPPINGS`) under which
// this plugin registers its mapper. It matches the name accepted by the historic
// hardcoded switch, so existing configuration keeps working unchanged.
const NameBodyToContext = "bodytocontext"

// init registers the BodyToContext mapper with the core registry, so that a
// blank/underscore import of this package makes it resolvable by name.
func init() {
	mapping.Register(NameBodyToContext, BodyToContext)
}
